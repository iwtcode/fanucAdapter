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
	"math"
	"strings"
	"unsafe"
)

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

// readDiagnosis — это внутренняя функция, которая выполняет основной вызов cnc_diagnoss
// и автоматически определяет тип возвращаемых данных (byte, word, dword)
func readDiagnosis(handle uint16, diagNo int16, axisNo int16, length int16) (int64, error) {
	buffer := make([]byte, 16)

	rc := C.go_cnc_diagnoss(
		C.ushort(handle),
		C.short(diagNo),
		C.short(axisNo),
		C.short(length),
		(*C.ODBDGN)(unsafe.Pointer(&buffer[0])),
	)

	if rc != C.EW_OK {
		return 0, fmt.Errorf("cnc_diagnoss for diagNo %d failed: rc=%d", diagNo, int16(rc))
	}

	// Определяем тип данных на основе поля `type` из ответа (первые 2 байта)
	// Старший байт поля `type` указывает на тип данных
	rawType := binary.LittleEndian.Uint16(buffer[0:2])
	dataType := rawType & 0xFF00

	// Данные всегда находятся со смещением 4 байта
	const dataOffset = 4

	switch dataType {
	case 0x0000: // Byte type
		return int64(buffer[dataOffset]), nil
	case 0x0100: // Word type (2 байта)
		val := int16(binary.LittleEndian.Uint16(buffer[dataOffset : dataOffset+2]))
		return int64(val), nil
	case 0x0200: // Double word type (4 байта)
		val := int32(binary.LittleEndian.Uint32(buffer[dataOffset : dataOffset+4]))
		return int64(val), nil
	default:
		return 0, fmt.Errorf("unknown data type for cnc_diagnoss %d, axis %d. Type: %d", diagNo, axisNo, rawType)
	}
}

// ReadDiagnosisByte считывает 1-байтовое диагностическое значение.
func ReadDiagnosisByte(handle uint16, diagNo int16, axisNo int16) (int32, error) {
	// length = 4 (header) + 1 (data)
	val, err := readDiagnosis(handle, diagNo, axisNo, 5)
	return int32(val), err
}

// ReadDiagnosisWord считывает 2-байтовое диагностическое значение.
func ReadDiagnosisWord(handle uint16, diagNo int16, axisNo int16) (int32, error) {
	// length = 4 (header) + 2 (data)
	val, err := readDiagnosis(handle, diagNo, axisNo, 6)
	return int32(val), err
}

// ReadDiagnosisDoubleWord считывает 4-байтовое диагностическое значение.
func ReadDiagnosisDoubleWord(handle uint16, diagNo int16, axisNo int16) (int64, error) {
	// length = 4 (header) + 4 (data)
	return readDiagnosis(handle, diagNo, axisNo, 8)
}

// ReadDiagnosisReal считывает 8-байтовое диагностическое значение с плавающей запятой.
// FANUC FOCAS представляет такие числа как два 4-байтовых целых:
// одно для значения (val), другое для положения десятичной точки (dec).
// Итоговое значение = val / (10^dec).
func ReadDiagnosisReal(handle uint16, diagNo int16, axisNo int16) (float64, error) {
	// Общая длина запроса для real-типа:
	// 4 байта (заголовок) + 8 байт (данные: 4 для значения + 4 для десятичной точки)
	const length = 12
	buffer := make([]byte, 16) // Буфер с запасом

	rc := C.go_cnc_diagnoss(
		C.ushort(handle),
		C.short(diagNo),
		C.short(axisNo),
		C.short(length),
		(*C.ODBDGN)(unsafe.Pointer(&buffer[0])),
	)

	if rc != C.EW_OK {
		return 0, fmt.Errorf("cnc_diagnoss for real diagNo %d (axis %d) failed: rc=%d", diagNo, axisNo, int16(rc))
	}

	// Данные всегда находятся со смещением 4 байта от начала буфера
	const dataOffset = 4

	// Первые 4 байта данных - это основное значение (dgn_val)
	val := int32(binary.LittleEndian.Uint32(buffer[dataOffset : dataOffset+4]))

	// Следующие 4 байта - это количество знаков после запятой (dec_val)
	decPos := int32(binary.LittleEndian.Uint32(buffer[dataOffset+4 : dataOffset+8]))

	// Рассчитываем итоговое значениеss
	realValue := float64(val) / math.Pow(10, float64(decPos))

	return realValue, nil
}
