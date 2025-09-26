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
	"unsafe"

	"github.com/iwtcode/fanucService/models"
)

// ReadAxisData считывает имена, абсолютные позиции и диагностику для всех управляемых осей
func ReadAxisData(handle uint16, numAxes int16, maxAxes int16) ([]models.AxisInfo, error) {
	if numAxes <= 0 {
		return []models.AxisInfo{}, nil
	}

	const odbposSize = 48
	bufferSize := int(maxAxes) * odbposSize
	buffer := make([]byte, bufferSize)
	axesToRead := C.short(maxAxes)

	rc := C.go_cnc_rdposition(C.ushort(handle), -1, &axesToRead, (*C.ODBPOS)(unsafe.Pointer(&buffer[0])))
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_rdposition failed: rc=%d", int16(rc))
	}

	if int(axesToRead) > int(maxAxes) {
		axesToRead = C.short(maxAxes)
	}
	if int(axesToRead) <= 0 {
		return []models.AxisInfo{}, nil
	}

	axisInfos := make([]models.AxisInfo, 0, axesToRead)

	for i := 0; i < int(axesToRead); i++ {
		offset := i * odbposSize

		posDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+0 : offset+4]))
		posDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+4 : offset+6]))
		posNameChar := buffer[offset+10]
		posSuffChar := buffer[offset+11]

		if posNameChar == 0 {
			continue
		}

		fullName := string(posNameChar)
		if posSuffChar != 0 && posSuffChar != ' ' {
			fullName += string(posSuffChar)
		}

		position := float64(posDataVal) / math.Pow(10, float64(posDecVal))

		axisNumber := int16(i + 1)

		var diag301Value float64
		if val, err := ReadDiagnosisReal(handle, 301, axisNumber); err == nil {
			diag301Value = val
		}

		var servoTempValue int32
		if val, err := ReadDiagnosisByte(handle, 308, axisNumber); err == nil {
			servoTempValue = val
		}

		var coderTempValue int32
		if val, err := ReadDiagnosisByte(handle, 309, axisNumber); err == nil {
			coderTempValue = val
		}

		var powerConsumptionValue int64
		if val, err := ReadDiagnosisDoubleWord(handle, 4901, axisNumber); err == nil {
			powerConsumptionValue = val
		}

		axisInfos = append(axisInfos, models.AxisInfo{
			Name:             trimNull(fullName),
			Position:         position,
			Diag301:          diag301Value,
			ServoTemperature: servoTempValue,
			CoderTemperature: coderTempValue,
			PowerConsumption: int32(powerConsumptionValue),
		})
	}

	return axisInfos, nil
}
