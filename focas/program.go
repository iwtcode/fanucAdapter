package focas

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../ -lfwlib32

#include <stdlib.h>
#include "c_helpers.h"
*/
import "C"

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/iwtcode/fanucService/models"
)

// ReadProgram считывает информацию о текущей выполняемой программе и текущую строку G-кода.
func (a *FocasAdapter) ReadProgram() (*models.ProgramInfo, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	var rc C.short

	err := a.callWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_exeprgname rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
	}

	progInfo := &models.ProgramInfo{
		Name:   trimNull(string(nameBuf)),
		Number: int64(onum),
	}

	var length C.ushort = 256
	var blknum C.short
	dataBuf := make([]byte, length)
	// Этот вызов можно не оборачивать, т.к. его ошибка не критична для получения основной информации
	rcExec := C.go_cnc_rdexecprog(C.ushort(a.handle), &length, &blknum, (*C.char)(unsafe.Pointer(&dataBuf[0])))

	if rcExec == C.EW_OK {
		fullBlock := trimNull(string(dataBuf[:length]))
		lines := strings.Split(fullBlock, "\n")
		progInfo.CurrentGCode = lines[0]
	} else {
		progInfo.CurrentGCode = ""
	}

	return progInfo, nil
}

// GetControlProgram считывает полное содержимое текущей выполняемой программы.
func (a *FocasAdapter) GetControlProgram() (string, error) {
	progInfo, err := a.ReadProgram()
	if err != nil {
		return "", fmt.Errorf("could not read program info: %w", err)
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
		return "", fmt.Errorf("no program is currently running or program number could not be determined (name: '%s', number: %d)", progInfo.Name, progInfo.Number)
	}

	// Оборачиваем всю последовательность загрузки в callWithReconnect
	var finalContent string
	err = a.callWithReconnect(func(handle uint16) (int16, error) {
		rc := C.go_cnc_upstart(C.ushort(handle), C.short(programNumberToUpload))
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_upstart for program '%s' (number %d) failed: rc=%d", progInfo.Name, programNumberToUpload, int16(rc))
		}
		defer C.go_cnc_upend(C.ushort(handle))

		var sb strings.Builder
		var buffer C.ODBUP

		for {
			length := C.ushort(256)
			rcUpload := C.go_cnc_upload(C.ushort(handle), &buffer, &length)

			if length > 0 {
				goBytes := C.GoBytes(unsafe.Pointer(&buffer.data[0]), C.int(length))
				sb.Write(goBytes)
			}

			if rcUpload != C.EW_OK && rcUpload != C.EW_BUFFER {
				break
			}
			if rcUpload == C.EW_OK && length == 0 {
				break
			}
		}

		rawContent := strings.ReplaceAll(sb.String(), "\x00", "")
		firstPercent := strings.Index(rawContent, "%")
		lastPercent := strings.LastIndex(rawContent, "%")

		if firstPercent != -1 && lastPercent > firstPercent {
			finalContent = rawContent[:lastPercent+1]
		} else {
			finalContent = rawContent
		}

		finalContent = strings.TrimSpace(finalContent)
		if !strings.HasPrefix(finalContent, "%") {
			finalContent = "%\n" + finalContent
		}

		return C.EW_OK, nil // Вся последовательность успешна
	})

	return finalContent, err
}
