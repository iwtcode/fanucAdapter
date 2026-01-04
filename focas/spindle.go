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

	"github.com/iwtcode/fanucAdapter/models"
)

// ReadSpindleData считывает информацию о скорости, нагрузке и коррекции для всех активных шпинделей.
func (a *FocasAdapter) ReadSpindleData() ([]models.SpindleInfo, error) {
	// 1. Чтение основных данных шпинделей (Load, Speed
	var numSpindles C.short = 8
	const sploadSpspeedSize = 24
	bufferSize := int(numSpindles) * sploadSpspeedSize
	buffer := make([]byte, bufferSize)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_rdspmeter(C.ushort(handle), -1, &numSpindles, (*C.ODBSPLOAD)(unsafe.Pointer(&buffer[0])))
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdspmeter failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
	}

	if numSpindles <= 0 {
		return []models.SpindleInfo{}, nil
	}

	// 2. Чтение Override (Коррекция)
	var overrideData C.ODBSPN
	errOverride := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc := C.go_cnc_rdspload(C.ushort(handle), -1, &overrideData)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("go_cnc_rdspload failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	// 3. Чтение диагностики 411
	maxAxes := int16(32)
	if a.sysInfo != nil && a.sysInfo.MaxAxes > 0 {
		maxAxes = a.sysInfo.MaxAxes
	}

	diag411Vals, errDiag := a.ReadDiagnosisWordAllAxes(411, maxAxes)
	if errDiag != nil {
		a.logger.Warnf("Warning: Batch read diag 411 failed: %v", errDiag)
		diag411Vals = make([]int32, maxAxes)
	}

	spindleInfos := make([]models.SpindleInfo, 0, numSpindles)

	for i := 0; i < int(numSpindles); i++ {
		offset := i * sploadSpspeedSize

		// Парсинг нагрузки
		loadDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+0 : offset+4]))
		loadDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+4 : offset+6]))
		load := float64(loadDataVal) / math.Pow(10, float64(loadDecVal))

		// Парсинг скорости
		speedDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+12 : offset+16]))
		speedDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+16 : offset+18]))
		rawSpeed := float64(speedDataVal) / math.Pow(10, float64(speedDecVal))

		// Коррекция скорости
		correctedSpeed := rawSpeed / 2.0
		speed := int32(math.Round(correctedSpeed))

		// Парсинг Override
		var overridePercent int16
		if errOverride == nil && i < len(overrideData.data) {
			rawOverride := overrideData.data[i]
			const maxOverrideValue = 16383.0
			calculatedPercent := (float64(rawOverride) / maxOverrideValue) * 100.0
			overridePercent = int16(math.Round(calculatedPercent))
		}

		// Получение значения диагностики 411 из предварительно загруженного массива.
		// Логика: SpindleNumber = i + 1.
		// Массив diag411Vals имеет индексацию с 0.
		// Предполагаем, что индекс оси шпинделя совпадает с порядковым номером (i).
		var diag411Value int32
		if i < len(diag411Vals) {
			diag411Value = diag411Vals[i]
		}

		spindleNumber := int16(i + 1)

		spindleInfos = append(spindleInfos, models.SpindleInfo{
			Number:          spindleNumber,
			SpeedRPM:        speed,
			LoadPercent:     load,
			OverridePercent: overridePercent,
			Diag411Value:    diag411Value,
		})
	}

	return spindleInfos, nil
}
