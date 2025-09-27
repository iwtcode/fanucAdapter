package program

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../ -lfwlib32

#include <stdlib.h>
#include "../c_helpers.h"
*/
import "C"
import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/iwtcode/fanucService/focas/model"
)

// ModelUnknownProgramReader предоставляет реализацию по умолчанию для чтения управляющей программы.
type ModelUnknownProgramReader struct{}

// GetControlProgram считывает полное содержимое текущей выполняемой программы.
func (pr *ModelUnknownProgramReader) GetControlProgram(a model.FocasCaller) (string, error) {
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
	log.Printf("Starting program upload for program O%d (%s)", programNumberToUpload, progInfo.Name)

	var finalContent string
	err = a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc := C.go_cnc_upstart(C.ushort(handle), C.short(programNumberToUpload))
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("cnc_upstart for program '%s' (number %d) failed: rc=%d", progInfo.Name, programNumberToUpload, int16(rc))
		}
		log.Printf("cnc_upstart successful for program O%d.", programNumberToUpload)
		defer C.go_cnc_upend(C.ushort(handle))

		var sb strings.Builder
		var uploadErr error
		var lastRc C.short

		const (
			EW_NODATA = -2 // Нет данных (конец файла)
			EW_BUSY   = -8 // Контроллер занят
		)

		iteration := 0
		for {
			var buffer C.ODBUP
			length := C.ushort(256)
			rcUpload := C.go_cnc_upload(C.ushort(handle), &buffer, &length)
			lastRc = rcUpload

			log.Printf("Upload iteration %d: rc=%d, length=%d", iteration, rcUpload, length)
			iteration++

			isDataRead := rcUpload == C.EW_OK || rcUpload == C.EW_BUFFER

			// Записываем данные в буфер только в случае успешного чтения
			if isDataRead && length > 0 {
				goBytes := C.GoBytes(unsafe.Pointer(&buffer.data[0]), C.int(length))
				sb.Write(goBytes)
			}

			// --- Логика управления циклом ---

			// 1. Условия успешного завершения
			if (rcUpload == C.EW_OK && length == 0) || rcUpload == EW_NODATA {
				log.Printf("Upload finished successfully with code: %d", rcUpload)
				break
			}

			// 2. Условие для повторной попытки
			if rcUpload == EW_BUSY {
				log.Printf("CNC is busy (rc=%d). Retrying in 50ms...", rcUpload)
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// 3. Условие продолжения чтения (буфер был полон, есть еще данные)
			if rcUpload == C.EW_BUFFER {
				continue
			}

			// 4. Условие неустранимой ошибки (все остальные случаи)
			if rcUpload != C.EW_OK {
				log.Printf("Exiting upload loop due to unrecoverable error. rc=%d", rcUpload)
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
		log.Printf("Total raw content size after loop: %d bytes", len(rawContent))

		// Надежно очищаем и обрамляем программу символами '%'
		trimmedContent := strings.Trim(rawContent, " \t\n\r%")
		finalContent = "%\n" + trimmedContent + "\n%"

		log.Printf("Final processed content size: %d bytes", len(finalContent))

		return C.EW_OK, nil // Вся последовательность успешна
	})

	if err != nil {
		return "", err // Если CallWithReconnect вернул ошибку, передаем ее дальше
	}

	return finalContent, nil
}
