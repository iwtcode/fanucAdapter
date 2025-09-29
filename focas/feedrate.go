package focas

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../ -lfwlib32

#include "c_helpers.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"unsafe"

	"github.com/iwtcode/fanucService/models"
)

// ReadFeedData считывает фактическую скорость подачи и процент коррекции.
// Реализация основана на C# коде, считывающем данные с помощью cnc_rdspeed и cnc_rdparam.
func (a *FocasAdapter) ReadFeedData() (*models.FeedInfo, error) {
	log.Println("[ReadFeedData] Начато чтение данных о скорости подачи и коррекции.")
	feedInfo := &models.FeedInfo{}
	var finalErr error

	// 1. Чтение фактической скорости подачи с помощью cnc_rdspeed
	// Размер структуры ODBSPEED примерно 32 байта.
	speedBuffer := make([]byte, 32)
	var rcSpeed C.short

	errSpeed := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rcSpeed = C.go_cnc_rdspeed(
			C.ushort(handle),
			0, // ИСПРАВЛЕНО: Тип 0 для фактической скорости подачи (был 2)
			(*C.ODBSPEED)(unsafe.Pointer(&speedBuffer[0])),
		)
		if rcSpeed != C.EW_OK {
			return int16(rcSpeed), fmt.Errorf("cnc_rdspeed failed: rc=%d", int16(rcSpeed))
		}
		return int16(rcSpeed), nil
	})

	if errSpeed != nil {
		log.Printf("[ReadFeedData] Ошибка при чтении фактической скорости подачи: %v.", errSpeed)
		finalErr = errSpeed // Сохраняем первую ошибку
	} else {
		// Парсинг буфера ODBSPEED. `actf` - первый член структуры.
		// Это структура REALDATA: long data (4 байта), short dec (2 байта).
		rateVal := int32(binary.LittleEndian.Uint32(speedBuffer[0:4]))
		rateDec := int16(binary.LittleEndian.Uint16(speedBuffer[4:6]))

		divisor := math.Pow(10, float64(rateDec))
		var actualFeedRate float64
		if divisor != 0 {
			actualFeedRate = float64(rateVal) / divisor
		}

		feedInfo.ActualFeedRate = int32(actualFeedRate)
		log.Printf("[ReadFeedData] Успешно прочитана фактическая скорость подачи. Значение: %d (Сырое: %d, Десятичные: %d)", feedInfo.ActualFeedRate, rateVal, rateDec)
	}

	// 2. Чтение коррекции подачи с помощью cnc_rdparam
	const paramNum = 20 // Номер параметра для коррекции подачи
	const axisNum = 0   // Номер оси (0 для общих параметров)
	const length = 8    // Длина структуры данных для одного параметра
	paramBuffer := make([]byte, length)
	var rcParam C.short

	errParam := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rcParam = C.go_cnc_rdparam(
			C.ushort(handle),
			C.short(paramNum),
			C.short(axisNum),
			C.short(length),
			(*C.IODBPSD)(unsafe.Pointer(&paramBuffer[0])),
		)
		if rcParam != C.EW_OK {
			return int16(rcParam), fmt.Errorf("cnc_rdparam для коррекции подачи (параметр %d) завершился с ошибкой: rc=%d", paramNum, int16(rcParam))
		}
		return int16(rcParam), nil
	})

	if errParam != nil {
		log.Printf("[ReadFeedData] Ошибка при чтении коррекции подачи: %v.", errParam)
		if finalErr == nil { // Не перезаписываем первую ошибку
			finalErr = errParam
		}
	} else {
		// Структура IODBPSD: short prm_no, short axis_no, затем union с данными.
		// Данные начинаются со смещения 4. Нас интересует значение типа short (idata).
		overrideValue := int16(binary.LittleEndian.Uint16(paramBuffer[4:6]))
		feedInfo.FeedOverride = overrideValue
		log.Printf("[ReadFeedData] Успешно прочитана коррекция подачи. Значение: %d", feedInfo.FeedOverride)
	}

	log.Printf("[ReadFeedData] Чтение завершено. Результат: %+v, Ошибка: %v", feedInfo, finalErr)
	return feedInfo, finalErr
}
