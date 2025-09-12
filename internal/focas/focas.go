package focas

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../../ -lfwlib32

#include <stdlib.h>
#include <stdint.h>
#include "c_helpers.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"github.com/iwtcode/fanucService/internal/domain"
)

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

// Startup инициализирует процесс FOCAS2
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

// Connect подключается к станку и возвращает хендл
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

// Disconnect освобождает хендл подключения
func Disconnect(handle uint16) {
	if handle == 0 {
		return
	}
	C.go_cnc_freelibhndl(C.ushort(handle))
}

// ReadProgram считывает информацию о текущей выполняемой программе
func ReadProgram(handle uint16) (*domain.ProgramInfo, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	rc := C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_exeprgname rc=%d", int16(rc))
	}
	return &domain.ProgramInfo{
		Name:   trimNull(string(nameBuf)),
		Number: int64(onum),
	}, nil
}

// ReadFullProgram считывает информацию и полное содержимое текущей выполняемой программы
func ReadFullProgram(handle uint16) (*domain.ControlProgram, error) {
	// 1. Получаем информацию о программе (имя и номер)
	progInfo, err := ReadProgram(handle)
	if err != nil {
		return nil, fmt.Errorf("could not read program info: %w", err)
	}

	var programNumberToUpload int64
	if strings.HasPrefix(progInfo.Name, "O") {
		parsedNum, err := strconv.ParseInt(strings.TrimSpace(progInfo.Name[1:]), 10, 64)
		if err == nil {
			programNumberToUpload = parsedNum
		}
	}

	if programNumberToUpload == 0 {
		programNumberToUpload = progInfo.Number
	}

	if programNumberToUpload <= 0 {
		return nil, fmt.Errorf("no program is currently running or program number could not be determined (name: '%s', number: %d)", progInfo.Name, progInfo.Number)
	}

	// 2. Начинаем процесс выгрузки программы с корректным номером
	rc := C.go_cnc_upstart(C.ushort(handle), C.short(programNumberToUpload))
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_upstart for program '%s' (number %d) failed: rc=%d", progInfo.Name, programNumberToUpload, int16(rc))
	}
	defer C.go_cnc_upend(C.ushort(handle))

	// 3. Читаем программу в цикле
	var sb strings.Builder
	var buffer C.ODBUP

	for {
		length := C.ushort(256)
		rcUpload := C.go_cnc_upload(C.ushort(handle), &buffer, &length)

		if length > 0 {
			goBytes := C.GoBytes(unsafe.Pointer(&buffer.data[0]), C.int(length))
			cleanBytes := bytes.ReplaceAll(goBytes, []byte{10}, []byte{10})
			sb.Write(bytes.TrimRight(cleanBytes, "\x00"))
		}

		// Любой код, кроме EW_OK, означает завершение передачи (успешное или нет).
		// Мы выходим из цикла, но не считаем это фатальной ошибкой, так как данные могли быть прочитаны.
		if rcUpload != C.EW_OK {
			break
		}
	}

	return &domain.ControlProgram{
		ProgramInfo: *progInfo,
		Content:     sb.String(),
	}, nil
}

// ReadSystemInfo считывает и возвращает системную информацию о станке
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
		controlledAxes = 0
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

// ReadAxisData считывает имена и абсолютные позиции для всех управляемых осей
func ReadAxisData(handle uint16, numAxes int16, maxAxes int16) ([]domain.AxisInfo, error) {
	if numAxes <= 0 {
		return []domain.AxisInfo{}, nil
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
		return []domain.AxisInfo{}, nil
	}

	axisInfos := make([]domain.AxisInfo, 0, axesToRead)

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

		axisInfos = append(axisInfos, domain.AxisInfo{
			Name:     trimNull(fullName),
			Position: position,
		})
	}

	return axisInfos, nil
}

// ReadMachineState считывает полное состояние станка и передает его интерпретатору
func ReadMachineState(handle uint16) (*domain.UnifiedMachineData, error) {
	var stat C.ODBST
	rc := C.go_cnc_statinfo(C.ushort(handle), &stat)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_statinfo rc=%d", int16(rc))
	}

	return InterpretMachineState(&stat), nil
}
