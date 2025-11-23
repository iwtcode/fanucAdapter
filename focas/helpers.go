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
	"math"
	"strings"
	"unsafe"
)

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

// readDiagnosisInternal - базовый метод для чтения диагностики
func (a *FocasAdapter) readDiagnosisInternal(diagNo int16, axisNo int16, length int16) ([]byte, error) {
	buffer := make([]byte, length)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
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
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// --- Методы для одной оси ---

func (a *FocasAdapter) ReadDiagnosisByte(diagNo int16, axisNo int16) (int32, error) {
	buf, err := a.readDiagnosisInternal(diagNo, axisNo, 5) // 4 header + 1 data
	if err != nil {
		return 0, err
	}
	return int32(buf[4]), nil
}

func (a *FocasAdapter) ReadDiagnosisWord(diagNo int16, axisNo int16) (int32, error) {
	buf, err := a.readDiagnosisInternal(diagNo, axisNo, 6) // 4 header + 2 data
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint16(buf[4:6])), nil
}

func (a *FocasAdapter) ReadDiagnosisDoubleWord(diagNo int16, axisNo int16) (int64, error) {
	buf, err := a.readDiagnosisInternal(diagNo, axisNo, 8) // 4 header + 4 data
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint32(buf[4:8])), nil
}

func (a *FocasAdapter) ReadDiagnosisReal(diagNo int16, axisNo int16) (float64, error) {
	buf, err := a.readDiagnosisInternal(diagNo, axisNo, 12) // 4 header + 8 data
	if err != nil {
		return 0, err
	}
	val := int32(binary.LittleEndian.Uint32(buf[4:8]))
	decPos := int32(binary.LittleEndian.Uint32(buf[8:12]))
	return float64(val) / math.Pow(10, float64(decPos)), nil
}

// --- Новые методы для массового чтения (ВСЕ ОСИ) ---

// ReadDiagnosisByteAllAxes читает байтовую диагностику для всех осей сразу.
func (a *FocasAdapter) ReadDiagnosisByteAllAxes(diagNo int16, maxAxes int16) ([]int32, error) {
	// Header (4 bytes) + (1 byte * maxAxes)
	length := int16(4 + 1*int(maxAxes))

	buf, err := a.readDiagnosisInternal(diagNo, -1, length)
	if err != nil {
		return nil, err
	}

	results := make([]int32, maxAxes)
	for i := 0; i < int(maxAxes); i++ {
		offset := 4 + i
		if offset < len(buf) {
			results[i] = int32(buf[offset])
		}
	}
	return results, nil
}

// ReadDiagnosisDoubleWordAllAxes читает 4-байтовую диагностику для всех осей.
func (a *FocasAdapter) ReadDiagnosisDoubleWordAllAxes(diagNo int16, maxAxes int16) ([]int64, error) {
	// Header (4 bytes) + (4 bytes * maxAxes)
	length := int16(4 + 4*int(maxAxes))

	buf, err := a.readDiagnosisInternal(diagNo, -1, length)
	if err != nil {
		return nil, err
	}

	results := make([]int64, maxAxes)
	for i := 0; i < int(maxAxes); i++ {
		offset := 4 + (i * 4)
		if offset+4 <= len(buf) {
			results[i] = int64(binary.LittleEndian.Uint32(buf[offset : offset+4]))
		}
	}
	return results, nil
}

// ReadDiagnosisRealAllAxes читает Real диагностику (значение + дес. точка) для всех осей.
func (a *FocasAdapter) ReadDiagnosisRealAllAxes(diagNo int16, maxAxes int16) ([]float64, error) {
	// Header (4 bytes) + (8 bytes * maxAxes) -> RealData = 4 bytes val + 4 bytes dec
	length := int16(4 + 8*int(maxAxes))

	buf, err := a.readDiagnosisInternal(diagNo, -1, length)
	if err != nil {
		return nil, err
	}

	results := make([]float64, maxAxes)
	for i := 0; i < int(maxAxes); i++ {
		offset := 4 + (i * 8)
		if offset+8 <= len(buf) {
			val := int32(binary.LittleEndian.Uint32(buf[offset : offset+4]))
			decPos := int32(binary.LittleEndian.Uint32(buf[offset+4 : offset+8]))

			if decPos > 20 || decPos < 0 {
				decPos = 0
			}
			divisor := math.Pow(10, float64(decPos))
			if divisor == 0 {
				divisor = 1
			}
			results[i] = float64(val) / divisor
		}
	}
	return results, nil
}
