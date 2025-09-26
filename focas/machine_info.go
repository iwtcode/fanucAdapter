package focas

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../ -lfwlib32

#include "c_helpers.h"
*/
import "C"

import (
	"fmt"
	"strconv"

	"github.com/iwtcode/fanucService/models"
)

// ReadSystemInfo считывает и возвращает системную информацию о станке (не метод адаптера, т.к. используется при создании).
func ReadSystemInfo(handle uint16) (*models.SystemInfo, error) {
	var sysInfo C.ODBSYS
	rc := C.go_cnc_sysinfo(C.ushort(handle), &sysInfo)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_sysinfo rc=%d", int16(rc))
	}

	series := C.GoStringN(&sysInfo.series[0], C.int(len(sysInfo.series)))
	version := C.GoStringN(&sysInfo.version[0], C.int(len(sysInfo.version)))
	axesStr := C.GoStringN(&sysInfo.axes[0], C.int(len(sysInfo.axes)))

	controlledAxes, err := strconv.Atoi(trimNull(axesStr))
	if err != nil {
		controlledAxes = 0
	}

	data := &models.SystemInfo{
		Manufacturer:   "FANUC",
		Series:         trimNull(series),
		Version:        trimNull(version),
		Model:          fmt.Sprintf("Series %s Version %s", trimNull(series), trimNull(version)),
		MaxAxes:        int16(sysInfo.max_axis),
		ControlledAxes: int16(controlledAxes),
	}

	return data, nil
}

// ReadMachineState считывает полное состояние станка.
func (a *FocasAdapter) ReadMachineState() (*models.UnifiedMachineData, error) {
	var stat C.ODBST
	var rc C.short

	err := a.callWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_statinfo(C.ushort(handle), &stat)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("go_cnc_statinfo rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
	}

	return InterpretMachineState(&stat), nil
}
