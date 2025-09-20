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

// ReadSpindleData считывает информацию о скорости, нагрузке и коррекции для всех активных шпинделей.
func ReadSpindleData(handle uint16) ([]domain.SpindleInfo, error) {
	// 1. Получаем данные о нагрузке и скорости для всех шпинделей
	var numSpindles C.short = 4 // Максимальное количество шпинделей для чтения
	const sploadSpspeedSize = 24
	bufferSize := int(numSpindles) * sploadSpspeedSize
	buffer := make([]byte, bufferSize)

	rc := C.go_cnc_rdspmeter(C.ushort(handle), -1, &numSpindles, (*C.ODBSPLOAD)(unsafe.Pointer(&buffer[0])))
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_rdspmeter failed: rc=%d", int16(rc))
	}

	if numSpindles <= 0 {
		return []domain.SpindleInfo{}, nil
	}

	// 2. Получаем данные о проценте коррекции (override)
	var overrideData C.ODBSPN
	rcOverride := C.go_cnc_rdspload(C.ushort(handle), -1, &overrideData)

	spindleInfos := make([]domain.SpindleInfo, 0, numSpindles)

	for i := 0; i < int(numSpindles); i++ {
		offset := i * sploadSpspeedSize

		// Парсим поле spload (нагрузка)
		loadDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+0 : offset+4]))
		loadDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+4 : offset+6]))
		load := float64(loadDataVal) / math.Pow(10, float64(loadDecVal))

		// Парсим поле spspeed (скорость)
		speedDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+12 : offset+16]))
		speedDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+16 : offset+18]))
		rawSpeed := float64(speedDataVal) / math.Pow(10, float64(speedDecVal))

		// Применяем коррекцию и округляем до ближайшего целого
		correctedSpeed := rawSpeed / 2.0
		speed := int32(math.Round(correctedSpeed))

		var overridePercent int16
		if rcOverride == C.EW_OK && i < len(overrideData.data) {
			rawOverride := overrideData.data[i]

			// !!! ВАЖНО: Сырое значение override - это число от 0 до 16383.
			// Масштабируем его до диапазона 0-100%.
			const maxOverrideValue = 16383.0
			calculatedPercent := (float64(rawOverride) / maxOverrideValue) * 100.0
			overridePercent = int16(math.Round(calculatedPercent))
		}

		spindleInfos = append(spindleInfos, domain.SpindleInfo{
			Number:          int16(i + 1),
			SpeedRPM:        speed,
			LoadPercent:     load,
			OverridePercent: overridePercent,
		})
	}

	return spindleInfos, nil
}
