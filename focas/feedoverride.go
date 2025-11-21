package focas

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -lfwlib32
#cgo linux LDFLAGS: -Wl,-rpath,${SRCDIR}

#include "c_helpers.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"log"
	"unsafe"
)

// ReadFeedOverride считывает процент коррекции подачи (F%).
// Используется FOCAS функция cnc_rdtofs.
func (a *FocasAdapter) ReadFeedOverride() (int32, error) {
	log.Println("[ReadFeedOverride] Начато чтение коррекции подачи с помощью cnc_rdtofs.")

	// Размер структуры ODBTOFS: datano(2) + type(2) + data(4) = 8 байт
	const dataLength = 8
	buffer := make([]byte, dataLength)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_rdtofs(
			C.ushort(handle),
			1, // номер 1 для коррекции подачи (F%)
			0, // тип 0
			C.short(dataLength),
			(*C.ODBTOFS)(unsafe.Pointer(&buffer[0])),
		)

		log.Printf("[ReadFeedOverride] Вызов go_cnc_rdtofs. Код возврата (rc): %d", rc)
		log.Printf("[ReadFeedOverride] Сырой буфер ответа (hex): %x", buffer)

		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdtofs failed with error code: %d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		log.Printf("[ReadFeedOverride] Ошибка после вызова CallWithReconnect: %v", err)
		return 0, err
	}

	// Данные (процент коррекции) находятся в поле `data` типа long,
	// которое начинается со смещения 4 байта в структуре ODBTOFS.
	overrideValue := int32(binary.LittleEndian.Uint32(buffer[4:8]))

	log.Printf("[ReadFeedOverride] Успешно прочитана коррекция подачи. Распарсенное значение: %d", overrideValue)
	return overrideValue, nil
}
