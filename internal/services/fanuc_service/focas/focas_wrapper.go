package focas

/*
#cgo CFLAGS: -I../../../../
// Linux/Unix:
#cgo LDFLAGS: -L../../../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// Windows (MinGW) вариант:
// #cgo windows LDFLAGS: -L../../../../ -lfwlib32

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

// Получение системной информации (ODBSYS)
static short go_cnc_sysinfo(unsigned short h, ODBSYS* sys_info_out) {
    return cnc_sysinfo(h, sys_info_out);
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/iwtcode/fanucService/internal/domain/models"
)

// InitializeFocas запускает FOCAS2 процесс. Должен вызываться один раз при старте приложения.
func InitializeFocas() error {
	logPath := os.Getenv("FOCAS_LOG_PATH")
	if logPath == "" {
		logPath = "./focas2.log"
	}
	dir := filepath.Dir(logPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	cpath := C.CString(logPath)
	defer C.free(unsafe.Pointer(cpath))

	rc := C.go_cnc_startupprocess(3, cpath)
	if rc != C.EW_OK {
		return fmt.Errorf("cnc_startupprocess(3, %q) failed with rc=%d", logPath, int16(rc))
	}
	return nil
}

// Connect пытается установить соединение с ЧПУ и возвращает хендл.
func Connect(ip string, port uint16, timeoutMs int32) (uint16, error) {
	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))

	var h C.ushort
	rc := C.go_cnc_allclibhndl3(cip, C.ushort(port), C.long(timeoutMs), &h)
	if rc != C.EW_OK {
		return 0, fmt.Errorf("cnc_allclibhndl3 for %s:%d failed: rc=%d", ip, port, int16(rc))
	}
	return uint16(h), nil
}

// Disconnect закрывает соединение.
func Disconnect(handle uint16) {
	if handle == 0 {
		return
	}
	C.go_cnc_freelibhndl(C.ushort(handle))
}

// GetAllData собирает всю доступную информацию со станка и преобразует ее в унифицированную модель для Kafka.
func GetAllData(handle uint16, sessionID, endpoint string) (*models.MachineDataKafka, error) {
	// Системная информация в текущей модели не используется, но получаем ее для полноты
	_, _ = readSystemInfo(handle)
	progInfo, _ := readProgram(handle)
	statusInfo, err := readMachineState(handle)
	if err != nil {
		return nil, fmt.Errorf("критическая ошибка: не удалось прочитать состояние станка: %w", err)
	}

	// --- Начало логики трансформации ---
	unavailable := "UNAVAILABLE"

	var currentProgram *models.CurrentProgramInfoKafka
	if progInfo != nil && (progInfo.Name != "" || progInfo.Number != 0) {
		currentProgram = &models.CurrentProgramInfoKafka{
			Program:        progInfo.Name,
			ProgramComment: fmt.Sprintf("O%d", progInfo.Number),
		}
	}

	alarms := make([]map[string]interface{}, 0)
	// Статус "Others" или "UNKNOWN" означает отсутствие активных алармов.
	hasAlarms := !(statusInfo.AlarmStatus == "Others" || statusInfo.AlarmStatus == "UNKNOWN" || statusInfo.AlarmStatus == "")

	if hasAlarms {
		alarm := make(map[string]interface{})
		alarm["level"] = "FAULT" // FOCAS не разделяет ошибки и предупреждения, используем общий уровень
		alarm["message"] = statusInfo.AlarmStatus
		alarms = append(alarms, alarm)
	}

	isManualMode := statusInfo.ProgramMode == "HaNDle" || statusInfo.ProgramMode == "JOG" || statusInfo.ProgramMode == "Teach in JOG" || statusInfo.ProgramMode == "Teach in HaNDle"

	machineData := &models.MachineDataKafka{
		MachineId:           endpoint,
		Id:                  "", // Поле 'id' пока не используется, как и запрошено
		Timestamp:           time.Now().Format(time.RFC3339),
		IsEnabled:           true, // Если соединение есть, считаем, что станок включен
		IsInEmergency:       statusInfo.EmergencyStatus == "EMerGency",
		MachineState:        statusInfo.MachineState,
		ProgramMode:         statusInfo.ProgramMode,
		TmMode:              statusInfo.TmMode,
		HandleRetraceStatus: unavailable,
		AxisMovementStatus:  statusInfo.AxisMovementStatus,
		MstbStatus:          statusInfo.MstbStatus,
		EmergencyStatus:     statusInfo.EmergencyStatus,
		AlarmStatus:         statusInfo.AlarmStatus,
		EditStatus:          statusInfo.EditStatus,
		ManualMode:          isManualMode,
		WriteStatus:         unavailable,
		LabelSkipStatus:     unavailable,
		WarningStatus:       "NORMAL",
		BatteryStatus:       statusInfo.AlarmStatus == "BATtery Low",
		ActiveToolNumber:    unavailable,
		ToolOffsetNumber:    unavailable,
		Alarms:              alarms,
		HasAlarms:           hasAlarms,
		PartsCount:          make(map[string]string),
		AccumulatedTime:     make(map[string]string),
		CurrentProgram:      currentProgram,
	}

	if hasAlarms {
		machineData.WarningStatus = "ACTIVE"
	}

	return machineData, nil
}

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

func readProgram(handle uint16) (*models.ProgramInfo, error) {
	nameBuf := make([]byte, 64)
	var onum C.long
	rc := C.go_cnc_exeprgname(C.ushort(handle), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.int(len(nameBuf)), &onum)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_exeprgname rc=%d", int16(rc))
	}
	return &models.ProgramInfo{
		Name:   trimNull(string(nameBuf)),
		Number: int64(onum),
	}, nil
}

func readSystemInfo(handle uint16) (*models.SystemInfo, error) {
	var sysInfo C.ODBSYS
	rc := C.go_cnc_sysinfo(C.ushort(handle), &sysInfo)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_sysinfo rc=%d", int16(rc))
	}

	series := C.GoStringN(&sysInfo.series[0], C.int(len(sysInfo.series)))
	version := C.GoStringN(&sysInfo.version[0], C.int(len(sysInfo.version)))

	data := &models.SystemInfo{
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
			return "APC"
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
			return "APC"
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
func readMachineState(handle uint16) (*models.UnifiedMachineData, error) {
	var stat C.ODBST
	rc := C.go_cnc_statinfo(C.ushort(handle), &stat)
	if rc != C.EW_OK {
		return nil, fmt.Errorf("go_cnc_statinfo rc=%d", int16(rc))
	}

	data := &models.UnifiedMachineData{}

	switch stat.tmmode {
	case 0:
		data.TmMode = "T"
	case 1:
		data.TmMode = "M"
	default:
		data.TmMode = "UNKNOWN"
	}

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
		data.MachineState = "MSTR"
	default:
		data.MachineState = "UNKNOWN"
	}

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

	if stat.mstb == 1 {
		data.MstbStatus = "FIN"
	} else {
		data.MstbStatus = "Other"
	}

	switch stat.emergency {
	case 0:
		data.EmergencyStatus = "Not Emergency"
	case 1:
		data.EmergencyStatus = "EMerGency"
	case 2:
		data.EmergencyStatus = "ReSET"
	case 3:
		data.EmergencyStatus = "WAIT"
	default:
		data.EmergencyStatus = "UNKNOWN"
	}

	switch stat.alarm {
	case 0:
		data.AlarmStatus = "Others"
	case 1:
		data.AlarmStatus = "ALarM"
	case 2:
		data.AlarmStatus = "BATtery Low"
	case 3:
		data.AlarmStatus = "FAN"
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

	data.EditStatus = interpretEditStatus(stat.tmmode, stat.edit)

	return data, nil
}
