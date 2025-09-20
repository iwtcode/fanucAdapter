package focas

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../../ -lfwlib32

#include <stdlib.h>
#include "c_helpers.h"
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

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
