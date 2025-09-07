package focas

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../../ -lfwlib32

#include <stdlib.h>
#include <stdint.h> // <--- ДОБАВЛЕНО
#include "c_helpers.h"
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe" // <--- ДОБАВЛЕНО

	"github.com/iwtcode/fanucService/pkg/domain"
)

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

// Startup инициализирует процесс FOCAS2.
func Startup(mode uint16, logPath string) error {
	dir := filepath.Dir(logPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	cpath := C.CString(logPath)
	defer C.free(unsafe.Pointer(cpath))

	rc := C.go_cnc_startupprocess(C.ushort(mode), cpath)
	if rc != C.EW_OK {
		return fmt.Errorf("cnc_startupprocess(%d, %q) rc=%d", mode, logPath, int16(rc))
	}
	return nil
}

// Connect подключается к станку и возвращает хендл.
func Connect(ip string, port uint16, timeoutMs int32) (uint16, error) {
	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))

	var h C.ushort
	rc := C.go_cnc_allclibhndl3(cip, C.ushort(port), C.long(timeoutMs), &h)
	if rc != C.EW_OK {
		return 0, fmt.Errorf("cnc_allclibhndl3 failed: rc=%d", int16(rc))
	}
	return uint16(h), nil
}

// Disconnect освобождает хендл подключения.
func Disconnect(handle uint16) {
	if handle == 0 {
		return
	}
	C.go_cnc_freelibhndl(C.ushort(handle))
}

// ReadProgram считывает информацию о текущей выполняемой программе.
func ReadProgram(handle uint16) (domain.ProgramInfo, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	rc := C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
	if rc != C.EW_OK {
		return domain.ProgramInfo{}, fmt.Errorf("cnc_exeprgname rc=%d", int16(rc))
	}
	return domain.ProgramInfo{
		Name:   trimNull(string(nameBuf)),
		Number: int64(onum),
	}, nil
}

// ReadSystemInfo считывает и возвращает системную информацию о станке.
func ReadSystemInfo(handle uint16) (*domain.SystemInfo, error) {
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
		controlledAxes = 0 // В случае ошибки парсинга, считаем что осей 0
	}

	data := &domain.SystemInfo{
		Manufacturer:   "FANUC",
		Series:         trimNull(series),
		Version:        trimNull(version),
		Model:          fmt.Sprintf("Series %s Version %s", trimNull(series), trimNull(version)),
		MaxAxis:        int16(sysInfo.max_axis),
		ControlledAxes: int16(controlledAxes),
	}

	return data, nil
}

// ReadAxisData считывает имена и абсолютные позиции для всех управляемых осей.
func ReadAxisData(handle uint16, numAxes int16, maxAxes int16) ([]domain.AxisInfo, error) {
	if numAxes <= 0 {
		return []domain.AxisInfo{}, nil
	}

	var axisPositions C.ODBAXIS

	length := C.short(4 + 4*C.MAX_AXIS)
	rcPos := C.go_cnc_absolute(C.ushort(handle), -1, length, &axisPositions)
	if rcPos != C.EW_OK {
		return nil, fmt.Errorf("cnc_absolute failed: rc=%d", int16(rcPos))
	}

	dataPtr := (*[C.MAX_AXIS]C.int32_t)(unsafe.Pointer(&axisPositions.data[0]))

	var axisName C.ODBAXISNAME
	axisInfos := make([]domain.AxisInfo, 0, numAxes)

	for i := int16(1); i <= numAxes; i++ {
		rcName := C.go_cnc_rdaxisname(C.ushort(handle), C.short(i), &axisName)
		if rcName != C.EW_OK {
			return nil, fmt.Errorf("cnc_rdaxisname for axis %d failed: rc=%d", i, int16(rcName))
		}

		nameChar := C.GoStringN(&axisName.name, 1)
		suffixChar := C.GoStringN(&axisName.suff, 1)

		fullName := nameChar
		if suffixChar[0] != 0 && suffixChar[0] != ' ' {
			fullName += suffixChar
		}

		position := int64(dataPtr[i-1])

		axisInfos = append(axisInfos, domain.AxisInfo{
			Name:     trimNull(fullName),
			Position: position,
		})
	}

	return axisInfos, nil
}

// ReadMachineState считывает полное состояние станка и передает его интерпретатору.
func ReadMachineState(handle uint16) (*domain.UnifiedMachineData, error) {
	var stat C.ODBST
	rc := C.go_cnc_statinfo(C.ushort(handle), &stat)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_statinfo rc=%d", int16(rc))
	}

	return InterpretMachineState(&stat), nil
}
