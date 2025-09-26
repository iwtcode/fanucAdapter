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

func (a *FocasAdapter) readDiagnosis(diagNo int16, axisNo int16, length int16) (int64, error) {
	buffer := make([]byte, 16)
	var rc C.short
	var rawType uint16

	err := a.callWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_diagnoss(
			C.ushort(handle),
			C.short(diagNo),
			C.short(axisNo),
			C.short(length),
			(*C.ODBDGN)(unsafe.Pointer(&buffer[0])),
		)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_diagnoss for diagNo %d failed: rc=%d", diagNo, int16(rc))
		}
		// Успешно, считываем тип данных
		rawType = binary.LittleEndian.Uint16(buffer[0:2])
		return int16(rc), nil
	})

	if err != nil {
		return 0, err
	}

	dataType := rawType & 0xFF00
	const dataOffset = 4

	switch dataType {
	case 0x0000: // Byte
		return int64(buffer[dataOffset]), nil
	case 0x0100: // Word
		val := int16(binary.LittleEndian.Uint16(buffer[dataOffset : dataOffset+2]))
		return int64(val), nil
	case 0x0200: // Double word
		val := int32(binary.LittleEndian.Uint32(buffer[dataOffset : dataOffset+4]))
		return int64(val), nil
	default:
		return 0, fmt.Errorf("unknown data type for cnc_diagnoss %d, axis %d. Type: %d", diagNo, axisNo, rawType)
	}
}

// ReadDiagnosisByte считывает 1-байтовое диагностическое значение.
func (a *FocasAdapter) ReadDiagnosisByte(diagNo int16, axisNo int16) (int32, error) {
	val, err := a.readDiagnosis(diagNo, axisNo, 5) // length = 4 (header) + 1 (data)
	return int32(val), err
}

// ReadDiagnosisWord считывает 2-байтовое диагностическое значение.
func (a *FocasAdapter) ReadDiagnosisWord(diagNo int16, axisNo int16) (int32, error) {
	val, err := a.readDiagnosis(diagNo, axisNo, 6) // length = 4 (header) + 2 (data)
	return int32(val), err
}

// ReadDiagnosisDoubleWord считывает 4-байтовое диагностическое значение.
func (a *FocasAdapter) ReadDiagnosisDoubleWord(diagNo int16, axisNo int16) (int64, error) {
	return a.readDiagnosis(diagNo, axisNo, 8) // length = 4 (header) + 4 (data)
}

// ReadDiagnosisReal считывает 8-байтовое диагностическое значение с плавающей запятой.
func (a *FocasAdapter) ReadDiagnosisReal(diagNo int16, axisNo int16) (float64, error) {
	const length = 12
	buffer := make([]byte, 16)
	var rc C.short

	err := a.callWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_diagnoss(
			C.ushort(handle),
			C.short(diagNo),
			C.short(axisNo),
			C.short(length),
			(*C.ODBDGN)(unsafe.Pointer(&buffer[0])),
		)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_diagnoss for real diagNo %d (axis %d) failed: rc=%d", diagNo, axisNo, int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return 0, err
	}

	const dataOffset = 4
	val := int32(binary.LittleEndian.Uint32(buffer[dataOffset : dataOffset+4]))
	decPos := int32(binary.LittleEndian.Uint32(buffer[dataOffset+4 : dataOffset+8]))

	realValue := float64(val) / math.Pow(10, float64(decPos))
	return realValue, nil
}
