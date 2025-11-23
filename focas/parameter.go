package focas

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/iwtcode/fanucAdapter/models"
)

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -lfwlib32
#cgo linux LDFLAGS: -Wl,-rpath,${SRCDIR}

#include "c_helpers.h"
*/
import "C"

const (
	paramPartsCount    = 6711 // Количество обработанных деталей
	paramPowerOnTime   = 6750 // Время включения
	paramOperatingTime = 6751 // Время работы
	paramCuttingTime   = 6753 // Время резания
	paramCycleTime     = 6757 // Время цикла
)

// formatDuration форматирует time.Duration в строку "HH:MM:SS".
func formatDuration(d time.Duration) string {
	totalSeconds := int64(d.Seconds())
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// ReadParameterInfo считывает и сразу форматирует группу параметров одним пакетным запросом.
func (a *FocasAdapter) ReadParameterInfo() (*models.ParameterInfo, error) {
	info := &models.ParameterInfo{}

	const startParam = 6711
	const endParam = 6757

	// Размер одного параметра (IODBPSD) для long целого = 8 байт
	// (2 байта datano + 2 байта type + 4 байта ldata)
	const paramSize = 8
	const bufferSize = 4096 // Достаточно для ~500 параметров
	buffer := make([]byte, bufferSize)

	var rc C.short
	var length C.short = C.short(bufferSize)
	var startNo C.short = C.short(startParam)
	var endNo C.short = C.short(endParam)

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_rdparar(
			C.ushort(handle),
			&startNo, // Указатель на start
			0,        // Axis
			&endNo,   // Указатель на end
			&length,  // Указатель на length
			(*C.IODBPSD)(unsafe.Pointer(&buffer[0])),
		)

		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_rdparar failed for range %d-%d: rc=%d", startParam, endParam, int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		log.Printf("Error reading parameters: %v", err)
		return info, err
	}

	bytesRead := int(length)
	offset := 0

	// Парсинг ответа
	for offset+paramSize <= bytesRead {
		// IODBPSD:
		// 0-2: datano (short)
		// 2-4: type (short)
		// 4-8: value (long/int32 - при условии type=INTEGER)

		prmNo := int16(binary.LittleEndian.Uint16(buffer[offset : offset+2]))
		val := int32(binary.LittleEndian.Uint32(buffer[offset+4 : offset+8]))

		switch prmNo {
		case paramPartsCount:
			info.PartsCount = int64(val)
		case paramPowerOnTime:
			duration := time.Duration(val) * time.Second
			info.PowerOnTime = formatDuration(duration)
		case paramOperatingTime:
			duration := time.Duration(val) * time.Second
			info.OperatingTime = formatDuration(duration)
		case paramCuttingTime:
			duration := time.Duration(val) * time.Second
			info.CuttingTime = formatDuration(duration)
		case paramCycleTime:
			duration := time.Duration(val) * time.Second
			info.CycleTime = formatDuration(duration)
		}

		offset += paramSize
	}

	return info, nil
}
