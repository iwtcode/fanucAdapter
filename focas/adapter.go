package focas

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../ -lfwlib32

#include <stdlib.h>
#include <string.h>
#include "c_helpers.h"
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/iwtcode/fanucService/focas/model"
	"github.com/iwtcode/fanucService/models"
)

// FocasAdapter инкапсулирует логику подключения и вызовов к FOCAS API.
// Он также управляет автоматическим переподключением и содержит реализации
// для конкретной модели станка.
type FocasAdapter struct {
	ip            string
	port          uint16
	timeout       int32
	handle        uint16
	mu            sync.Mutex
	sysInfo       *models.SystemInfo
	interpreter   model.Interpreter   // Интерфейс для интерпретации состояния
	programReader model.ProgramReader // Интерфейс для чтения программы
}

// Убедимся, что FocasAdapter удовлетворяет интерфейсу FocasCaller.
var _ model.FocasCaller = (*FocasAdapter)(nil)

// NewFocasAdapter создает новый экземпляр FocasAdapter и устанавливает соединение.
func NewFocasAdapter(ip string, port uint16, timeoutMs int32, modelSeries string) (*FocasAdapter, error) {
	handle, err := Connect(ip, port, timeoutMs)
	if err != nil {
		return nil, fmt.Errorf("initial connection failed: %w", err)
	}

	// Используем фабрику с вручную указанной серией для получения нужных реализаций
	interpreter, programReader := GetModelImplementations(modelSeries)

	adapter := &FocasAdapter{
		ip:            ip,
		port:          port,
		timeout:       timeoutMs,
		handle:        handle,
		interpreter:   interpreter,
		programReader: programReader,
	}

	sysInfo, err := adapter.ReadSystemInfo()
	if err != nil {
		Disconnect(handle)
		return nil, fmt.Errorf("failed to read system info after connecting: %w", err)
	}
	adapter.sysInfo = sysInfo

	return adapter, nil
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

// Reconnect пытается восстановить соединение.
func (a *FocasAdapter) Reconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.handle != 0 {
		Disconnect(a.handle)
		a.handle = 0
		time.Sleep(50 * time.Millisecond)
	}

	newHandle, err := Connect(a.ip, a.port, a.timeout)
	if err != nil {
		return fmt.Errorf("Reconnect failed: %w", err)
	}

	a.handle = newHandle
	fmt.Println("Successfully reconnected to FOCAS.")
	return nil
}

// CallWithReconnect — это обертка для выполнения вызовов с возможностью переподключения.
func (a *FocasAdapter) CallWithReconnect(f func(handle uint16) (int16, error)) error {
	for {
		a.mu.Lock()
		currentHandle := a.handle
		a.mu.Unlock()

		rc, err := f(currentHandle)

		if err == nil {
			return nil
		}

		if rc == C.EW_HANDLE || rc == C.EW_SOCKET {
			fmt.Printf("Connection error detected (rc=%d). Attempting to Reconnect...\n", rc)

			if reconnErr := a.Reconnect(); reconnErr != nil {
				fmt.Printf("Reconnect failed: %v. Retrying in 1 second...\n", reconnErr)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			continue

		} else {
			return err
		}
	}
}

// Close закрывает соединение.
func (a *FocasAdapter) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.handle != 0 {
		Disconnect(a.handle)
		a.handle = 0
	}
}

// GetSystemInfo возвращает системную информацию о станке.
func (a *FocasAdapter) GetSystemInfo() *models.SystemInfo {
	return a.sysInfo
}

// ReadSystemInfo считывает и возвращает системную информацию о станке.
func (a *FocasAdapter) ReadSystemInfo() (*models.SystemInfo, error) {
	var sysInfo C.ODBSYS
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_sysinfo(C.ushort(handle), &sysInfo)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("go_cnc_sysinfo rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
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

// ReadMachineState считывает и интерпретирует состояние станка, используя реализацию для конкретной модели.
func (a *FocasAdapter) ReadMachineState() (*models.UnifiedMachineData, error) {
	var stat C.ODBST
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
		rc = C.go_cnc_statinfo(C.ushort(handle), &stat)
		if rc != C.EW_OK {
			return int16(rc), fmt.Errorf("go_cnc_statinfo rc=%d", int16(rc))
		}
		return int16(rc), nil
	})

	if err != nil {
		return nil, err
	}

	// Делегируем интерпретацию состояния конкретной реализации, передавая указатель
	return a.interpreter.InterpretMachineState(unsafe.Pointer(&stat)), nil
}

// GetControlProgram считывает G-код программы, используя реализацию для конкретной модели.
func (a *FocasAdapter) GetControlProgram() (string, error) {
	// Делегируем чтение программы конкретной реализации, передавая ей себя
	// в качестве FocasCaller для выполнения низкоуровневых вызовов.
	return a.programReader.GetControlProgram(a)
}

// ReadProgram считывает информацию о текущей выполняемой программе и текущую строку G-кода.
// Этот метод является частью интерфейса model.FocasCaller.
func (a *FocasAdapter) ReadProgram() (*models.ProgramInfo, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	var rc C.short

	err := a.CallWithReconnect(func(handle uint16) (int16, error) {
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

	a.mu.Lock()
	currentHandle := a.handle
	a.mu.Unlock()
	rcExec := C.go_cnc_rdexecprog(C.ushort(currentHandle), &length, &blknum, (*C.char)(unsafe.Pointer(&dataBuf[0])))

	if rcExec == C.EW_OK {
		fullBlock := trimNull(string(dataBuf[:length]))
		lines := strings.Split(fullBlock, "\n")
		progInfo.CurrentGCode = lines[0]
	} else {
		progInfo.CurrentGCode = ""
	}

	return progInfo, nil
}
