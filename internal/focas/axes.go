package focas

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../../ -lfwlib32

#include "c_helpers.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/iwtcode/fanucService/internal/domain"
)

// ReadAxisData считывает имена и абсолютные позиции для всех управляемых осей
func ReadAxisData(handle uint16, numAxes int16, maxAxes int16) ([]domain.AxisInfo, error) {
	if numAxes <= 0 {
		return []domain.AxisInfo{}, nil
	}

	// Размер структуры ODBPOS равен 48 байтам (4 * POSELM, где POSELM = 12 байт)
	const odbposSize = 48
	bufferSize := int(maxAxes) * odbposSize
	buffer := make([]byte, bufferSize)
	axesToRead := C.short(maxAxes)

	// Вызываем C-функцию, передавая сырой байтовый буфер
	rc := C.go_cnc_rdposition(C.ushort(handle), -1, &axesToRead, (*C.ODBPOS)(unsafe.Pointer(&buffer[0])))
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_rdposition failed: rc=%d", int16(rc))
	}

	// axesToRead содержит фактическое количество прочитанных осей
	if int(axesToRead) > int(maxAxes) {
		axesToRead = C.short(maxAxes) // Не доверяем, если вернулось больше, чем мы просили
	}
	if int(axesToRead) <= 0 {
		return []domain.AxisInfo{}, nil // Нет данных
	}

	axisInfos := make([]domain.AxisInfo, 0, axesToRead)

	for i := 0; i < int(axesToRead); i++ {
		// Смещение для текущей структуры ODBPOS в буфере
		offset := i * odbposSize

		// Нас интересует только структура POSELM для абсолютной позиции,
		// которая находится в начале каждой ODBPOS (смещение 0).
		// Схема POSELM: data(4), dec(2), unit(2), disp(2), name(1), suff(1) = 12 байт

		// Читаем поля POSELM вручную из среза байтов, предполагая порядок LittleEndian
		// long data; (смещение 0, 4 байта)
		posDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+0 : offset+4]))
		// short dec; (смещение 4, 2 байта)
		posDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+4 : offset+6]))
		// char name; (смещение 10, 1 байт)
		posNameChar := buffer[offset+10]
		// char suff; (смещение 11, 1 байт)
		posSuffChar := buffer[offset+11]

		// Пропускаем оси без имени
		if posNameChar == 0 {
			continue
		}

		fullName := string(posNameChar)
		if posSuffChar != 0 && posSuffChar != ' ' {
			fullName += string(posSuffChar)
		}

		// Рассчитываем позицию с учетом десятичного разделителя
		position := float64(posDataVal) / math.Pow(10, float64(posDecVal))

		axisInfos = append(axisInfos, domain.AxisInfo{
			Name:     trimNull(fullName),
			Position: position,
		})
	}

	return axisInfos, nil
}
