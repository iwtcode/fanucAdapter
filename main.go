package main

/*
#cgo CFLAGS: -I.
// Если у вас только 32-битная libfwlib32.so — собирайте весь бинарь под 32 бита через окружение (GOARCH=386 и т.п.), а НЕ через #cgo.
// Linux/Unix:
#cgo LDFLAGS: -L. -lfwlib32 -Wl,-rpath,'$ORIGIN'
// Windows (MinGW) вариант — при необходимости раскомментируйте:
// #cgo windows LDFLAGS: -L. -lfwlib32

#include <stdlib.h>
#include <string.h>
#include "fwlib32.h"

// ---- C-helpers ----

// Инициализация процесса FOCAS2 и файл лога
static short go_cnc_startupprocess(unsigned short mode, const char* logpath) {
    return cnc_startupprocess(mode, logpath);
}

// Прокси для cnc_allclibhndl3
static short go_cnc_allclibhndl3(const char* ip, unsigned short port, long timeout_ms, unsigned short* handle_out) {
    return cnc_allclibhndl3(ip, port, timeout_ms, handle_out);
}

// Прокси для cnc_freelibhndl
static short go_cnc_freelibhndl(unsigned short h) {
    return cnc_freelibhndl(h);
}

// Программа: имя и номер (ODBEXEPRG.name / .o_num)
static short go_cnc_exeprgname(unsigned short h, char* name_out, int name_cap, long* onum_out) {
    ODBEXEPRG p;
    short rc = cnc_exeprgname(h, &p);
    if (rc == EW_OK) {
        int n = (int)sizeof(p.name);
        if (n >= name_cap) n = name_cap - 1;
        memcpy(name_out, p.name, n);
        name_out[n] = '\0';
        *onum_out = p.o_num;
    }
    return rc;
}

// Получение полной информации о статусе станка (ODBST)
static short go_cnc_statinfo(unsigned short h, ODBST* stat_out) {
	return cnc_statinfo(h, stat_out);
}

// ---- НОВЫЙ C-HELPER ДЛЯ ИНФОРМАЦИИ О СИСТЕМЕ ----
// Получение системной информации (ODBSYS)
static short go_cnc_sysinfo(unsigned short h, ODBSYS* sys_info_out) {
    return cnc_sysinfo(h, sys_info_out);
}
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

// ProgramInfo содержит информацию о выполняемой программе.
type ProgramInfo struct {
	Name   string
	Number int64
}

// SystemInfo содержит системную информацию о станке.
type SystemInfo struct {
	Manufacturer string
	Model        string
	Series       string
	Version      string
}

// UnifiedMachineData содержит полное унифицированное состояние станка.
type UnifiedMachineData struct {
	TmMode             string
	ProgramMode        string
	MachineState       string
	AxisMovementStatus string
	MstbStatus         string
	EmergencyStatus    string
	AlarmStatus        string
	EditStatus         string
}

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

func focasStartup(mode uint16, logPath string) error {
	dir := filepath.Dir(logPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	cpath := C.CString(logPath)
	defer C.free(unsafe.Pointer(cpath))

	rc := C.go_cnc_startupprocess(C.ushort(mode), cpath)
	if rc != 0 {
		return fmt.Errorf("cnc_startupprocess(%d, %q) rc=%d", mode, logPath, int16(rc))
	}
	return nil
}

func connect(ip string, port uint16, timeoutMs int32) (uint16, error) {
	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))

	var h C.ushort
	rc := C.go_cnc_allclibhndl3(cip, C.ushort(port), C.long(timeoutMs), &h)
	if rc != 0 {
		return 0, fmt.Errorf("cnc_allclibhndl3 failed: rc=%d", int16(rc))
	}
	return uint16(h), nil
}

func disconnect(handle uint16) {
	if handle == 0 {
		return
	}
	C.go_cnc_freelibhndl(C.ushort(handle))
}

func readProgram(handle uint16) (ProgramInfo, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	rc := C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
	if rc != 0 {
		return ProgramInfo{}, fmt.Errorf("cnc_exeprgname rc=%d", int16(rc))
	}
	return ProgramInfo{
		Name:   trimNull(string(nameBuf)),
		Number: int64(onum),
	}, nil
}

// readSystemInfo считывает и возвращает системную информацию о станке.
func readSystemInfo(handle uint16) (*SystemInfo, error) {
	var sysInfo C.ODBSYS
	rc := C.go_cnc_sysinfo(C.ushort(handle), &sysInfo)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_sysinfo rc=%d", int16(rc))
	}

	series := C.GoStringN(&sysInfo.series[0], C.int(len(sysInfo.series)))
	version := C.GoStringN(&sysInfo.version[0], C.int(len(sysInfo.version)))

	data := &SystemInfo{
		Manufacturer: "FANUC",
		Series:       trimNull(series),
		Version:      trimNull(version),
		Model:        fmt.Sprintf("Series %s Version %s", trimNull(series), trimNull(version)),
	}

	return data, nil
}

// interpretEditStatus интерпретирует статус редактирования в зависимости от режима станка (T/M).
func interpretEditStatus(tmmode C.short, editValue C.short) string {
	switch tmmode {
	case 0: // T mode (токарный станок)
		switch editValue {
		case 0:
			return "Not Editing"
		case 1:
			return "EDIT"
		case 2:
			return "SEARCH"
		case 3:
			return "OUTPUT"
		case 4:
			return "INPUT"
		case 5:
			return "COMPARE"
		case 6:
			return "OFFSET"
		case 7:
			return "Work Shift"
		case 9:
			return "Restart"
		case 10:
			return "RVRS"
		case 11:
			return "RTRY"
		case 12:
			return "RVED"
		case 14:
			return "PTRR"
		case 16:
			return "AICC"
		case 21:
			return "HPCC"
		case 23:
			return "NANO HP"
		case 25:
			return "5-AXIS"
		case 26:
			return "OFSX"
		case 27:
			return "OFSZ"
		case 28:
			return "WZR"
		case 29:
			return "OFSY"
		case 31:
			return "TOFS"
		case 39:
			return "TCP"
		case 40:
			return "TWP"
		case 41:
			return "TCP+TWP"
		case 42:
			return "APC"
		case 43:
			return "PRG-CHK"
		case 44:
			return "APC" // Дублируется, но оставлено для соответствия
		case 45:
			return "S-TCP"
		case 59:
			return "ALLSAVE"
		case 60:
			return "NOTSAVE"
		default:
			return "UNKNOWN"
		}
	case 1: // M mode (фрезерный станок)
		switch editValue {
		case 0:
			return "Not Editing"
		case 1:
			return "EDIT"
		case 2:
			return "SEARCH"
		case 3:
			return "OUTPUT"
		case 4:
			return "INPUT"
		case 5:
			return "COMPARE"
		case 6:
			return "Label Skip"
		case 7:
			return "Restart"
		case 8:
			return "HPCC"
		case 9:
			return "PTRR"
		case 10:
			return "RVRS"
		case 11:
			return "RTRY"
		case 12:
			return "RVED"
		case 13:
			return "HANDLE"
		case 14:
			return "OFFSET"
		case 15:
			return "Work Offset"
		case 16:
			return "AICC"
		case 17:
			return "Memory Check"
		case 21:
			return "AI APC"
		case 22:
			return "MBL APC"
		case 23:
			return "NANO HP"
		case 24:
			return "AI HPCC"
		case 25:
			return "5-AXIS"
		case 26:
			return "LEN"
		case 27:
			return "RAD"
		case 28:
			return "WZR"
		case 39:
			return "TCP"
		case 40:
			return "TWP"
		case 41:
			return "TCP+TWP"
		case 42:
			return "APC"
		case 43:
			return "PRG-CHK"
		case 44:
			return "APC" // Дублируется, но оставлено для соответствия
		case 45:
			return "S-TCP"
		case 59:
			return "ALLSAVE"
		case 60:
			return "NOTSAVE"
		default:
			return "UNKNOWN"
		}
	default:
		return "UNKNOWN"
	}
}

// readMachineState считывает, интерпретирует и возвращает полное состояние станка.
func readMachineState(handle uint16) (*UnifiedMachineData, error) {
	var stat C.ODBST
	rc := C.go_cnc_statinfo(C.ushort(handle), &stat)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_statinfo rc=%d", int16(rc))
	}

	data := &UnifiedMachineData{}

	// 1. Режим T/M (tmmode)
	switch stat.tmmode {
	case 0:
		data.TmMode = "T" // Токарный
	case 1:
		data.TmMode = "M" // Фрезерный
	default:
		data.TmMode = "UNKNOWN"
	}

	// 2. Режим работы (aut)
	switch stat.aut {
	case 0:
		data.ProgramMode = "MDI"
	case 1:
		data.ProgramMode = "MEMory"
	case 2:
		data.ProgramMode = "No Selection"
	case 3:
		data.ProgramMode = "EDIT"
	case 4:
		data.ProgramMode = "HaNDle"
	case 5:
		data.ProgramMode = "JOG"
	case 6:
		data.ProgramMode = "Teach in JOG"
	case 7:
		data.ProgramMode = "Teach in HaNDle"
	case 8:
		data.ProgramMode = "INC·feed"
	case 9:
		data.ProgramMode = "REFerence"
	case 10:
		data.ProgramMode = "ReMoTe"
	default:
		data.ProgramMode = "UNKNOWN"
	}

	// 3. Статус выполнения программы (run)
	switch stat.run {
	case 0:
		data.MachineState = "Reset"
	case 1:
		data.MachineState = "STOP"
	case 2:
		data.MachineState = "HOLD"
	case 3:
		data.MachineState = "START"
	case 4:
		data.MachineState = "MSTR (during retraction and re-positioning of tool retraction and recovery, and operation of JOG MDI)"
	default:
		data.MachineState = "UNKNOWN"
	}

	// 4. Движение осей (motion)
	switch stat.motion {
	case 0:
		data.AxisMovementStatus = "None"
	case 1:
		data.AxisMovementStatus = "Motion"
	case 2:
		data.AxisMovementStatus = "Dwell"
	default:
		data.AxisMovementStatus = "UNKNOWN"
	}

	// 5. Статус M/S/T/B (mstb)
	if stat.mstb == 1 {
		data.MstbStatus = "FIN"
	} else {
		data.MstbStatus = "Other"
	}

	// 6. Аварийный стоп (emergency)
	switch stat.emergency {
	case 0:
		data.EmergencyStatus = "Not Emergency"
	case 1:
		data.EmergencyStatus = "EMerGency"
	case 2:
		data.EmergencyStatus = "ReSET"
	case 3:
		data.EmergencyStatus = "WAIT (FS35i only)"
	default:
		data.EmergencyStatus = "UNKNOWN"
	}

	// 7. Статус тревоги (alarm)
	switch stat.alarm {
	case 0:
		data.AlarmStatus = "Others"
	case 1:
		data.AlarmStatus = "ALarM"
	case 2:
		data.AlarmStatus = "BATtery Low"
	case 3:
		data.AlarmStatus = "FAN (NC or Servo amplifier)"
	case 4:
		data.AlarmStatus = "PS Warning"
	case 5:
		data.AlarmStatus = "FSsB Warning"
	case 6:
		data.AlarmStatus = "INSulate Warning"
	case 7:
		data.AlarmStatus = "ENCoder Warning"
	case 8:
		data.AlarmStatus = "PMC Alarm"
	default:
		data.AlarmStatus = "UNKNOWN"
	}

	// 8. Статус редактирования (edit)
	data.EditStatus = interpretEditStatus(stat.tmmode, stat.edit)

	return data, nil
}

// _______________________________________________________

func main() {
	// 1) Инициализация FOCAS2 процесса + лог
	logPath := os.Getenv("FOCAS_LOG_PATH")
	if logPath == "" {
		logPath = "./focas2.log"
	}
	if err := focasStartup(3, logPath); err != nil {
		log.Fatalf("FOCAS startup error: %v", err)
	}

	// 2) Подключение к ЧПУ
	ip := os.Getenv("FANUC_IP")
	if ip == "" {
		ip = "192.168.30.142"
	}
	port := uint16(8193)
	timeoutMs := int32(5000)

	fmt.Printf("Подключение к %s:%d ...\n", ip, port)
	h, err := connect(ip, port, timeoutMs)
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer disconnect(h)
	fmt.Println("Успешно подключено!")

	// 3) Чтение системной информации
	fmt.Println("\n--- Системная информация ---")
	systemInfo, err := readSystemInfo(h)
	if err != nil {
		log.Fatalf("Не удалось прочитать системную информацию: %v", err)
	}
	fmt.Printf("Производитель: %s\n", systemInfo.Manufacturer)
	fmt.Printf("Модель:        %s\n", systemInfo.Model)
	fmt.Printf("Серия:         %s\n", systemInfo.Series)
	fmt.Printf("Версия:        %s\n", systemInfo.Version)

	// 4) Чтение информации о программе
	fmt.Println("\n--- Информация о программе ---")
	progInfo, err := readProgram(h)
	if err != nil {
		log.Printf("Не удалось прочитать информацию о программе: %v", err)
	} else {
		fmt.Printf("Имя программы: %s\n", progInfo.Name)
		fmt.Printf("Номер программы: %d\n", progInfo.Number)
	}

	// 5) Чтение полного состояния станка
	fmt.Println("\n--- Состояние станка ---")
	machineData, err := readMachineState(h)
	if err != nil {
		log.Fatalf("Не удалось прочитать состояние станка: %v", err)
	}

	fmt.Printf("Режим (T/M):           %s\n", machineData.TmMode)
	fmt.Printf("Режим работы:          %s\n", machineData.ProgramMode)
	fmt.Printf("Состояние выполнения:    %s\n", machineData.MachineState)
	fmt.Printf("Движение осей:           %s\n", machineData.AxisMovementStatus)
	fmt.Printf("Статус M/S/T/B:          %s\n", machineData.MstbStatus)
	fmt.Printf("Статус аварийного стопа: %s\n", machineData.EmergencyStatus)
	fmt.Printf("Статус тревоги:          %s\n", machineData.AlarmStatus)
	fmt.Printf("Статус редактирования:   %s\n", machineData.EditStatus)
}
