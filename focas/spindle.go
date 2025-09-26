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

// ReadSpindleData считывает информацию о скорости, нагрузке и коррекции для всех активных шпинделей.
func (a *FocasAdapter) ReadSpindleData() ([]models.SpindleInfo, error) {
	var numSpindles C.short = 8
	const sploadSpspeedSize = 24
	bufferSize := int(numSpindles) * sploadSpspeedSize
	buffer := make([]byte, bufferSize)
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) { // ИСПРАВЛЕНО
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

	var overrideData C.ODBSPN
	a.mu.Lock()
	currentHandle := a.handle
	a.mu.Unlock()
	rcOverride := C.go_cnc_rdspload(C.ushort(currentHandle), -1, &overrideData) // Этот вызов можно не оборачивать, если он не критичен

	spindleInfos := make([]models.SpindleInfo, 0, numSpindles)

	for i := 0; i < int(numSpindles); i++ {
		offset := i * sploadSpspeedSize

		loadDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+0 : offset+4]))
		loadDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+4 : offset+6]))
		load := float64(loadDataVal) / math.Pow(10, float64(loadDecVal))

		speedDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+12 : offset+16]))
		speedDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+16 : offset+18]))
		rawSpeed := float64(speedDataVal) / math.Pow(10, float64(speedDecVal))

		correctedSpeed := rawSpeed / 2.0
		speed := int32(math.Round(correctedSpeed))

		var overridePercent int16
		if rcOverride == C.EW_OK && i < len(overrideData.data) {
			rawOverride := overrideData.data[i]
			const maxOverrideValue = 16383.0
			calculatedPercent := (float64(rawOverride) / maxOverrideValue) * 100.0
			overridePercent = int16(math.Round(calculatedPercent))
		}

		var diag411Value int32
		spindleNumber := int16(i + 1)
		if val, err := a.ReadDiagnosisWord(411, spindleNumber); err == nil {
			diag411Value = val
		}

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
