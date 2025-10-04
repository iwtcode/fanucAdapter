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
	"unsafe"
)

// ReadJogOverride считывает процент коррекции скорости перемещения в режиме JOG.
func (a *FocasAdapter) ReadJogOverride() (int32, error) {
	log.Println("[ReadJogOverride] Начато чтение коррекции JOG.")
	const length = 8 // Размер структуры ODBTOFS
	buffer := make([]byte, length)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_rdtofs(
			C.ushort(handle),
			1, // номер корректора
			1, // тип корректора
			C.short(length),
			(*C.ODBTOFS)(unsafe.Pointer(&buffer[0])),
		)

		log.Printf("[ReadFeedOverride] Вызов go_cnc_rdtofs. Код возврата (rc): %d", rc)
		log.Printf("[ReadFeedOverride] Сырой буфер ответа (hex): %x", buffer)

		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdtofs for JOG override failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		log.Printf("[ReadJogOverride] Ошибка при чтении коррекции JOG: %v.", err)
		return 0, err
	}

	// Структура ODBTOFS: short datano (2), short type (2), long data (4).
	// Нас интересует поле data, которое находится со смещением 4.
	jogOverride := int32(binary.LittleEndian.Uint32(buffer[4:8]))
	log.Printf("[ReadJogOverride] Успешно прочитана коррекция JOG. Значение: %d", jogOverride)

	return jogOverride, nil
}
