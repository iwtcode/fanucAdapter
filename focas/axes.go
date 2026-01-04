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
	"unsafe"

	. "github.com/iwtcode/fanucAdapter/focas/errcode"
	"github.com/iwtcode/fanucAdapter/models"
)

// ReadAxisData считывает имена, абсолютные позиции и диагностику для всех управляемых осей
func (a *FocasAdapter) ReadAxisData() ([]models.AxisInfo, error) {
	sysInfo := a.sysInfo
	if sysInfo == nil || sysInfo.ControlledAxes <= 0 {
		return []models.AxisInfo{}, nil
	}

	// Используем MaxAxes из sysinfo (в вашем случае 32) для расчета буферов
	maxAxes := sysInfo.MaxAxes
	const odbposSize = 48
	bufferSize := int(maxAxes) * odbposSize
	buffer := make([]byte, bufferSize)
	axesToRead := C.short(maxAxes)
	var rc C.short

	// 1. Читаем позиции (стандартный метод)
	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_rdposition(C.ushort(handle), -1, &axesToRead, (*C.ODBPOS)(unsafe.Pointer(&buffer[0])))
		if int16(rc) != EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdposition failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
	}

	// axesToRead возвращает реальное количество прочитанных осей (например, 3)
	// Но для диагностики нам нужно знать системный максимум (32).

	if int(axesToRead) <= 0 {
		return []models.AxisInfo{}, nil
	}

	// 2. Массовое чтение диагностики (OPTIMIZATION)
	// Передаем maxAxes (32), чтобы FOCAS не вернул ошибку длины.

	// Diag 301: Servo Load (Real)
	diag301Vals, err := a.ReadDiagnosisRealAllAxes(301, maxAxes)
	if err != nil {
		a.logger.Warnf("Warning: Batch read diag 301 failed: %v", err)
		diag301Vals = make([]float64, maxAxes)
	}

	// Diag 308: Servo Temperature (Byte)
	diag308Vals, err := a.ReadDiagnosisByteAllAxes(308, maxAxes)
	if err != nil {
		a.logger.Warnf("Warning: Batch read diag 308 failed: %v", err)
		diag308Vals = make([]int32, maxAxes)
	}

	// Diag 309: Coder Temperature (Byte)
	diag309Vals, err := a.ReadDiagnosisByteAllAxes(309, maxAxes)
	if err != nil {
		a.logger.Warnf("Warning: Batch read diag 309 failed: %v", err)
		diag309Vals = make([]int32, maxAxes)
	}

	// Diag 4901: Power Consumption (Double Word)
	diag4901Vals, err := a.ReadDiagnosisDoubleWordAllAxes(4901, maxAxes)
	if err != nil {
		// Это нормально для старых станков
		diag4901Vals = make([]int64, maxAxes)
	}

	axisInfos := make([]models.AxisInfo, 0, axesToRead)

	for i := 0; i < int(axesToRead); i++ {
		offset := i * odbposSize

		// Парсинг позиции из ODBPOS
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

		// Берем значения из массивов по индексу оси
		var d301 float64
		var d308 int32
		var d309 int32
		var d4901 int64

		if i < len(diag301Vals) {
			d301 = diag301Vals[i]
		}
		if i < len(diag308Vals) {
			d308 = diag308Vals[i]
		}
		if i < len(diag309Vals) {
			d309 = diag309Vals[i]
		}
		if i < len(diag4901Vals) {
			d4901 = diag4901Vals[i]
		}

		axisInfos = append(axisInfos, models.AxisInfo{
			Name:             trimNull(fullName),
			Position:         position,
			Diag301:          d301,
			ServoTemperature: d308,
			CoderTemperature: d309,
			PowerConsumption: int32(d4901),
		})
	}

	return axisInfos, nil
}
