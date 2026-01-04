package program

/*
#cgo CFLAGS: -I${SRCDIR}/..
#cgo LDFLAGS: -L${SRCDIR}/.. -lfwlib32
#cgo linux LDFLAGS: -Wl,-rpath,${SRCDIR}/..

#include <stdlib.h>
#include "../c_helpers.h"
*/
import "C"
import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"

	. "github.com/iwtcode/fanucAdapter/focas/errcode"
	"github.com/iwtcode/fanucAdapter/focas/model"
)

// ModelUnknownProgramReader предоставляет реализацию по умолчанию для чтения управляющей программы.
type ModelUnknownProgramReader struct{}

// GetControlProgram считывает полное содержимое текущей выполняемой программы.
func (pr *ModelUnknownProgramReader) GetControlProgram(a model.FocasCaller) (string, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	var progName string
	logger := a.Logger()

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc := C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
		if int16(rc) != EW_OK {
			return int16(rc), fmt.Errorf("cnc_exeprgname failed: rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return "", fmt.Errorf("could not read program info: %w", err)
	}

	progName = strings.TrimRight(string(nameBuf), "\x00")

	var programNumberToUpload int64
	if strings.HasPrefix(progName, "O") {
		parsedNum, err := strconv.ParseInt(strings.TrimSpace(progName[1:]), 10, 64)
		if err == nil {
			programNumberToUpload = parsedNum
		}
	}

	var finalContent string
	err = a.CallWithReconnect(func(handle uint16) (int16, error) {
		var rc C.short

		if programNumberToUpload > 0 {
			logger.Infof("Starting program upload by number for program O%d (%s)", programNumberToUpload, progName)
			rc = C.go_cnc_upstart(C.ushort(handle), C.short(programNumberToUpload))
			if int16(rc) != EW_OK {
				return int16(rc), fmt.Errorf("cnc_upstart for program '%s' (number %d) failed: rc=%d", progName, programNumberToUpload, int16(rc))
			}
		} else {
			var pathNo, maxPathNo C.short
			rcPath := C.go_cnc_getpath(C.ushort(handle), &pathNo, &maxPathNo)
			if int16(rcPath) != EW_OK {
				return int16(rcPath), fmt.Errorf("cnc_getpath failed: rc=%d", int16(rcPath))
			}

			filePath := fmt.Sprintf("//CNC_MEM/USER/PATH%d/%s", pathNo, progName)
			logger.Infof("Starting program upload by path for program '%s'", filePath)

			cFilePath := C.CString(filePath)
			defer C.free(unsafe.Pointer(cFilePath))

			rc = C.go_cnc_upstart4(C.ushort(handle), 0, cFilePath)
			if int16(rc) != EW_OK {
				return int16(rc), fmt.Errorf("cnc_upstart4 for program '%s' failed: rc=%d", filePath, int16(rc))
			}
		}

		logger.Info("Program upload started successfully.")
		defer C.go_cnc_upend(C.ushort(handle))

		var sb strings.Builder
		var uploadErr error
		var lastRc C.short

		iteration := 0
		for {
			var buffer C.ODBUP
			length := C.ushort(256)
			rcUpload := C.go_cnc_upload(C.ushort(handle), &buffer, &length)
			lastRc = rcUpload

			logger.Debugf("Upload iteration %d: rc=%d, length=%d", iteration, rcUpload, length)
			iteration++

			isDataRead := int16(rcUpload) == EW_OK || int16(rcUpload) == EW_BUFFER

			if isDataRead && length > 0 {
				goBytes := C.GoBytes(unsafe.Pointer(&buffer.data[0]), C.int(length))
				sb.Write(goBytes)
			}

			// 1. Условия успешного завершения
			if (int16(rcUpload) == EW_OK && length == 0) || int16(rcUpload) == EW_RESET {
				logger.Infof("Upload finished successfully with code: %d", rcUpload)
				break
			}

			// 2. Условие для повторной попытки
			if int16(rcUpload) == EW_HANDLE {
				logger.Warnf("CNC is busy (rc=%d). Retrying in 50ms...", rcUpload)
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// 3. Условие продолжения чтения (буфер был полон, есть еще данные)
			if int16(rcUpload) == EW_BUFFER {
				continue
			}

			// 4. Условие неустранимой ошибки (все остальные случаи)
			if int16(rcUpload) != EW_OK {
				logger.Errorf("Exiting upload loop due to unrecoverable error. rc=%d", rcUpload)
				uploadErr = fmt.Errorf("cnc_upload failed with rc=%d", int16(rcUpload))
				break
			}
		}

		// Если во время цикла произошла ошибка, немедленно возвращаем ее
		if uploadErr != nil {
			return int16(lastRc), uploadErr
		}

		// Обработка и очистка происходят только после успешной загрузки
		rawContent := strings.ReplaceAll(sb.String(), "\x00", "")
		logger.Debugf("Total raw content size after loop: %d bytes", len(rawContent))

		// Надежно очищаем и обрамляем программу символами '%'
		trimmedContent := strings.Trim(rawContent, " \t\n\r%")
		finalContent = "%\n" + trimmedContent + "\n%"

		logger.Debugf("Final processed content size: %d bytes", len(finalContent))

		return int16(EW_OK), nil // Вся последовательность успешна
	})

	if err != nil {
		return "", err // Если CallWithReconnect вернул ошибку, передаем ее дальше
	}

	return finalContent, nil
}
