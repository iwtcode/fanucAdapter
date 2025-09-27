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
	"strconv"
	"strings"
	"unsafe"

	"github.com/iwtcode/fanucService/focas/interpreter"
	"github.com/iwtcode/fanucService/models"
)

// ReadAlarms считывает все активные сообщения об ошибках со станка.
func (a *FocasAdapter) ReadAlarms() ([]models.AlarmDetail, error) {
	const maxAlarms = 10
	// Размер структуры ODBALMMSG2_data: alm_no(4) + type(2) + axis(2) + dummy(2) + msg_len(2) + alm_msg(64) = 76 байт
	const alarmDataSize = 76
	bufferSize := maxAlarms * alarmDataSize
	buffer := make([]byte, bufferSize)
	numAlarms := C.short(maxAlarms)
	var rc C.short

	log.Printf("[ReadAlarms] Попытка чтения до %d ошибок (структура ODBALMMSG2)...", maxAlarms)

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		log.Printf("[ReadAlarms] Вызов C.go_cnc_rdalmmsg с хендлом %d", handle)
		rc = C.go_cnc_rdalmmsg(
			C.ushort(handle),
			-1, // Читать все типы ошибок
			&numAlarms,
			(*C.ODBALMMSG)(unsafe.Pointer(&buffer[0])),
		)
		log.Printf("[ReadAlarms] C.go_cnc_rdalmmsg вернул: rc=%d, numAlarms=%d", rc, numAlarms)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdalmmsg failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		log.Printf("[ReadAlarms] Ошибка во время CallWithReconnect: %v", err)
		return nil, err
	}

	if numAlarms <= 0 {
		log.Println("[ReadAlarms] На станке не найдено активных ошибок.")
		return []models.AlarmDetail{}, nil
	}

	log.Printf("[ReadAlarms] Найдено ошибок: %d. Начинаю парсинг...", numAlarms)
	alarms := make([]models.AlarmDetail, 0, numAlarms)
	for i := 0; i < int(numAlarms); i++ {
		offset := i * alarmDataSize
		alarmBytes := buffer[offset : offset+alarmDataSize]
		log.Printf("[ReadAlarms] Обработка ошибки #%d, сырые байты: %x", i+1, alarmBytes)

		// Номер ошибки (long, 4 байта)
		alarmNumber := int32(binary.LittleEndian.Uint32(alarmBytes[0:4]))

		// Тип ошибки (short, 2 байта)
		alarmType := int16(binary.LittleEndian.Uint16(alarmBytes[4:6]))

		// Длина сообщения (short, 2 байта, смещение 10)
		msgLen := int(binary.LittleEndian.Uint16(alarmBytes[10:12]))

		log.Printf("[ReadAlarms] Распарсенные детали: Номер=%d, Тип=%d, ДлинаСообщения=%d", alarmNumber, alarmType, msgLen)

		// Сообщение (смещение 12, длина msgLen)
		var message string
		if msgLen > 0 {
			msgEnd := 12 + msgLen
			if msgEnd > len(alarmBytes) {
				msgEnd = len(alarmBytes) // Предохранитель от выхода за пределы
			}
			message = string(alarmBytes[12:msgEnd])
			message = strings.TrimSpace(message)
		}

		log.Printf("[ReadAlarms] Распарсенное сообщение: '%s'", message)

		if message == "" {
			log.Printf("[ReadAlarms] Пропуск ошибки #%d, так как сообщение пустое.", i+1)
			continue
		}

		alarmDetail := models.AlarmDetail{
			ErrorCode:            strconv.FormatInt(int64(alarmNumber), 10),
			ErrorMessage:         message,
			ErrorTypeDescription: getAlarmTypeDescription(alarmType),
		}
		alarms = append(alarms, alarmDetail)
	}

	log.Printf("[ReadAlarms] Парсинг завершен. Возвращаю %d ошибок.", len(alarms))
	return alarms, nil
}

// getAlarmTypeDescription преобразует код типа ошибки в человекочитаемое описание.
func getAlarmTypeDescription(alarmType int16) string {
	switch alarmType {
	case 0:
		return interpreter.AlarmTypeSW
	case 1:
		return interpreter.AlarmTypePW
	case 2:
		return interpreter.AlarmTypeIO
	case 3:
		return interpreter.AlarmTypePS
	case 4:
		return interpreter.AlarmTypeOT
	case 5:
		return interpreter.AlarmTypeOH
	case 6:
		return interpreter.AlarmTypeSV
	case 7:
		return interpreter.AlarmTypeSR
	case 8:
		return interpreter.AlarmTypeMC
	case 9:
		return interpreter.AlarmTypeSP
	case 10:
		return interpreter.AlarmTypeDS
	case 11:
		return interpreter.AlarmTypeIE
	case 12:
		return interpreter.AlarmTypeBG
	case 13:
		return interpreter.AlarmTypeSN
	case 14:
		return interpreter.AlarmReserved
	case 15:
		return interpreter.AlarmTypeEX
	case 19:
		return interpreter.AlarmTypePC
	default:
		return interpreter.StatusUnknown
	}
}
