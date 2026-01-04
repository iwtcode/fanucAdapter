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
	"unsafe"
)

// ReadContourFeedRate считывает фактическую скорость подачи по контуру (F).
// Эта функция вызывает cnc_actf.
func (a *FocasAdapter) ReadContourFeedRate() (int32, error) {
	a.logger.Debug("[ReadContourFeedRate] Начато чтение скорости подачи по контуру.")

	// Размер структуры ODBACT = 2 * short (4 байта) + 1 * long (4 байта) = 8 байт
	const dataSize = 8
	buffer := make([]byte, dataSize)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_actf(C.ushort(handle), (*C.ODBACT)(unsafe.Pointer(&buffer[0])))
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_actf failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		a.logger.Errorf("[ReadContourFeedRate] Ошибка при чтении скорости подачи по контуру: %v", err)
		return 0, err
	}

	// Логируем сырые байты, полученные от FOCAS
	a.logger.Debugf("[ReadContourFeedRate] Получены сырые байты: %x", buffer)

	// Ручное декодирование данных для надежности.
	// Поле `data` (long) находится по смещению 4 байта после `dummy[2]`.
	// Используем LittleEndian, так как это стандарт для FOCAS на x86.
	const dataOffset = 4
	if len(buffer) < dataOffset+4 {
		err := fmt.Errorf("ожидался буфер размером >= 8, но получен %d", len(buffer))
		a.logger.Errorf("[ReadContourFeedRate] Ошибка декодирования: %v", err)
		return 0, err
	}

	contourFeedRate := int32(binary.LittleEndian.Uint32(buffer[dataOffset : dataOffset+4]))

	a.logger.Debugf("[ReadContourFeedRate] Успешно прочитана скорость подачи по контуру. Значение: %d", contourFeedRate)
	return contourFeedRate, nil
}
