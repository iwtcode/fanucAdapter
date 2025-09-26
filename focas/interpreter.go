package focas

import (
	"github.com/iwtcode/fanucService/models"
)

/*
#include "fwlib32.h"
*/
import "C"

// InterpretMachineState принимает сырую структуру ODBST
// и преобразует ее в доменную модель UnifiedMachineData.
func InterpretMachineState(stat *C.ODBST) *models.UnifiedMachineData {
	return &models.UnifiedMachineData{
		TmMode:             interpretTmMode(stat.tmmode),
		ProgramMode:        interpretProgramMode(stat.aut),
		MachineState:       interpretMachineState(stat.run),
		AxisMovementStatus: interpretAxisMovement(stat.motion),
		MstbStatus:         interpretMstbStatus(stat.mstb),
		EmergencyStatus:    interpretEmergencyStatus(stat.emergency),
		AlarmStatus:        interpretAlarmStatus(stat.alarm),
		EditStatus:         interpretEditStatus(stat.tmmode, stat.edit),
	}
}

func interpretTmMode(tmmode C.short) string {
	switch tmmode {
	case 0:
		return "T" // Токарный
	case 1:
		return "M" // Фрезерный
	default:
		return "UNKNOWN"
	}
}

func interpretProgramMode(aut C.short) string {
	switch aut {
	case 0:
		return "MDI"
	case 1:
		return "MEMory"
	case 2:
		return "No Selection"
	case 3:
		return "EDIT"
	case 4:
		return "HaNDle"
	case 5:
		return "JOG"
	case 6:
		return "Teach in JOG"
	case 7:
		return "Teach in HaNDle"
	case 8:
		return "INC·feed"
	case 9:
		return "REFerence"
	case 10:
		return "ReMoTe"
	default:
		return "UNKNOWN"
	}
}

func interpretMachineState(run C.short) string {
	switch run {
	case 0:
		return "Reset"
	case 1:
		return "STOP"
	case 2:
		return "HOLD"
	case 3:
		return "START"
	case 4:
		return "MSTR (during retraction and re-positioning of tool retraction and recovery, and operation of JOG MDI)"
	default:
		return "UNKNOWN"
	}
}

func interpretAxisMovement(motion C.short) string {
	switch motion {
	case 0:
		return "None"
	case 1:
		return "Motion"
	case 2:
		return "Dwell"
	default:
		return "UNKNOWN"
	}
}

func interpretMstbStatus(mstb C.short) string {
	if mstb == 1 {
		return "FIN"
	}
	return "Other"
}

func interpretEmergencyStatus(emergency C.short) string {
	switch emergency {
	case 0:
		return "Not Emergency"
	case 1:
		return "EMerGency"
	case 2:
		return "ReSET"
	case 3:
		return "WAIT (FS35i only)"
	default:
		return "UNKNOWN"
	}
}

func interpretAlarmStatus(alarm C.short) string {
	switch alarm {
	case 0:
		return "Others"
	case 1:
		return "ALarM"
	case 2:
		return "BATtery Low"
	case 3:
		return "FAN (NC or Servo amplifier)"
	case 4:
		return "PS Warning"
	case 5:
		return "FSsB Warning"
	case 6:
		return "INSulate Warning"
	case 7:
		return "ENCoder Warning"
	case 8:
		return "PMC Alarm"
	default:
		return "UNKNOWN"
	}
}

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
