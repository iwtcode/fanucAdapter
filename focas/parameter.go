// D:\vs\go\fanucService\focas\parameter.go

package focas

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/iwtcode/fanucService/models"
)

/*
#include "c_helpers.h"
*/
import "C"

const (
	paramPartsCount    = 6711 // Количество обработанных деталей
	paramPowerOnTime   = 6750 // Время включения
	paramOperatingTime = 6751 // Время работы
	paramCuttingTime   = 6753 // Время резания
	paramCycleTime     = 6757 // Время цикла
)

// formatDuration форматирует time.Duration в строку "HH:MM:SS".
func formatDuration(d time.Duration) string {
	totalSeconds := int64(d.Seconds())
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// readParameterLong считывает 4-байтовый (long) параметр.
func (a *FocasAdapter) readParameterLong(prmNo int16) (int32, error) {
	const length = 8
	buffer := make([]byte, length)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_rdparam(
			C.ushort(handle),
			C.short(prmNo),
			0,
			C.short(length),
			(*C.IODBPSD)(unsafe.Pointer(&buffer[0])),
		)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdparam для параметра %d завершился с ошибкой: rc=%d", prmNo, int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return 0, err
	}

	// Данные ldata начинаются со смещения 4 (после datano и type)
	value := int32(binary.LittleEndian.Uint32(buffer[4:8]))
	return value, nil
}

// ReadParameterInfo считывает и сразу форматирует группу параметров.
func (a *FocasAdapter) ReadParameterInfo() (*models.ParameterInfo, error) {
	info := &models.ParameterInfo{}
	var firstErr error

	// Чтение количества деталей
	partsCount, err := a.readParameterLong(paramPartsCount)
	if err != nil {
		log.Printf("Warning: не удалось прочитать количество деталей (параметр %d): %v", paramPartsCount, err)
		firstErr = err
	} else {
		info.PartsCount = int64(partsCount)
	}

	// Чтение времени включения
	powerOnSeconds, err := a.readParameterLong(paramPowerOnTime)
	if err != nil {
		log.Printf("Warning: не удалось прочитать время включения (параметр %d): %v", paramPowerOnTime, err)
		if firstErr == nil {
			firstErr = err
		}
	} else {
		duration := time.Duration(powerOnSeconds) * time.Second
		info.PowerOnTime = formatDuration(duration)
	}

	// Чтение времени работы
	operatingSeconds, err := a.readParameterLong(paramOperatingTime)
	if err != nil {
		log.Printf("Warning: не удалось прочитать время работы (параметр %d): %v", paramOperatingTime, err)
		if firstErr == nil {
			firstErr = err
		}
	} else {
		duration := time.Duration(operatingSeconds) * time.Second
		info.OperatingTime = formatDuration(duration)
	}

	// Чтение времени резания
	cuttingSeconds, err := a.readParameterLong(paramCuttingTime)
	if err != nil {
		log.Printf("Warning: не удалось прочитать время резания (параметр %d): %v", paramCuttingTime, err)
		if firstErr == nil {
			firstErr = err
		}
	} else {
		duration := time.Duration(cuttingSeconds) * time.Second
		info.CuttingTime = formatDuration(duration)
	}

	// Чтение времени цикла
	cycleSeconds, err := a.readParameterLong(paramCycleTime)
	if err != nil {
		log.Printf("Warning: не удалось прочитать время цикла (параметр %d): %v", paramCycleTime, err)
		if firstErr == nil {
			firstErr = err
		}
	} else {
		duration := time.Duration(cycleSeconds) * time.Second
		info.CycleTime = formatDuration(duration)
	}

	return info, firstErr
}
