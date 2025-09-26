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

// ReadProgram считывает информацию о текущей выполняемой программе и текущую строку G-кода
func ReadProgram(handle uint16) (*models.ProgramInfo, error) {
	// 1. Получаем имя и номер программы
	nameBuf := make([]byte, 64)
	var onum C.long
	rc := C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_exeprgname rc=%d", int16(rc))
	}

	progInfo := &models.ProgramInfo{
		Name:   trimNull(string(nameBuf)),
		Number: int64(onum),
	}

	// 2. Получаем текущую выполняемую строку G-кода
	var length C.ushort = 256
	var blknum C.short
	dataBuf := make([]byte, length)
	rcExec := C.go_cnc_rdexecprog(C.ushort(handle), &length, &blknum, (*C.char)(unsafe.Pointer(&dataBuf[0])))

	if rcExec == C.EW_OK {
		// Функция может возвращать несколько блоков, разделенных символом новой строки.
		// Нас интересует только первая строка, которая является текущей выполняемой.
		fullBlock := trimNull(string(dataBuf[:length]))
		lines := strings.Split(fullBlock, "\n")
		if len(lines) > 0 {
			progInfo.CurrentGCode = lines[0]
		} else {
			progInfo.CurrentGCode = ""
		}
	} else {
		// Не возвращаем ошибку, так как информация о программе может быть полезна,
		// а строка G-кода может быть недоступна (например, если программа не выполняется).
		// Оставляем поле пустым.
		progInfo.CurrentGCode = ""
	}

	return progInfo, nil
}

// GetControlProgram считывает полное содержимое текущей выполняемой программы
func GetControlProgram(handle uint16) (string, error) {
	// 1. Получаем информацию о программе (имя и номер)
	progInfo, err := ReadProgram(handle)
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

	// 2. Начинаем процесс выгрузки программы с корректным номером
	rc := C.go_cnc_upstart(C.ushort(handle), C.short(programNumberToUpload))
	if rc != C.EW_OK {
		return "", fmt.Errorf("cnc_upstart for program '%s' (number %d) failed: rc=%d", progInfo.Name, programNumberToUpload, int16(rc))
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
			sb.Write(goBytes)
		}

		if rcUpload != C.EW_OK && rcUpload != C.EW_BUFFER {
			break
		}

		if rcUpload == C.EW_OK && length == 0 {
			break
		}
	}

	// 4. Обрабатываем собранные данные
	rawContent := strings.ReplaceAll(sb.String(), "\x00", "")
	firstPercent := strings.Index(rawContent, "%")
	lastPercent := strings.LastIndex(rawContent, "%")
	var finalContent string

	if firstPercent != -1 && lastPercent > firstPercent {
		finalContent = rawContent[:lastPercent+1]
	} else {
		finalContent = rawContent
	}

	finalContent = strings.TrimSpace(finalContent)

	if !strings.HasPrefix(finalContent, "%") {
		finalContent = "%\n" + finalContent
	}

	return finalContent, nil
}
