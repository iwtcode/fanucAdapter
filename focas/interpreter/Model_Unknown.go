package interpreter

import (
	"unsafe"

	"github.com/iwtcode/fanucService/models"
)

/*
#include "../fwlib32.h"
*/
import "C"

const (
	// TmMode (Тип станка)
	TmModeTurning = "T" // Токарный
	TmModeMilling = "M" // Фрезерный

	// ProgramMode (Режим программы)
	ProgramModeMDI           = "MDI"
	ProgramModeMemory        = "MEMory"
	ProgramModeNoSelection   = "No Selection"
	ProgramModeEdit          = "EDIT"
	ProgramModeHandle        = "HaNDle"
	ProgramModeJOG           = "JOG"
	ProgramModeTeachInJOG    = "Teach in JOG"
	ProgramModeTeachInHandle = "Teach in HaNDle"
	ProgramModeIncFeed       = "INC·feed"
	ProgramModeReference     = "REFerence"
	ProgramModeRemote        = "ReMoTe"

	// MachineState (Состояние станка)
	MachineStateReset = "Reset"
	MachineStateStop  = "STOP"
	MachineStateHold  = "HOLD"
	MachineStateStart = "START"
	MachineStateMSTR  = "MSTR (during retraction and re-positioning of tool retraction and recovery, and operation of JOG MDI)"

	// AxisMovement (Движение осей)
	AxisMovementNone   = "None"
	AxisMovementMotion = "Motion"
	AxisMovementDwell  = "Dwell"

	// MstbStatus
	MstbStatusFIN   = "FIN"
	MstbStatusOther = "Other"

	// EmergencyStatus (Статус аварийной остановки)
	EmergencyStatusNotEmergency = "Not Emergency"
	EmergencyStatusEmergency    = "EMerGency"
	EmergencyStatusReset        = "ReSET"
	EmergencyStatusWait         = "WAIT (FS35i only)"

	// AlarmStatus (Статус тревоги)
	AlarmStatusOthers          = "Others"
	AlarmStatusAlarm           = "ALarM"
	AlarmStatusBatteryLow      = "BATtery Low"
	AlarmStatusFan             = "FAN (NC or Servo amplifier)"
	AlarmStatusPSWarning       = "PS Warning"
	AlarmStatusFSSBWarning     = "FSsB Warning"
	AlarmStatusInsulateWarning = "INSulate Warning"
	AlarmStatusEncoderWarning  = "ENCoder Warning"
	AlarmStatusPMCAlarm        = "PMC Alarm"

	// EditStatus (Общие статусы редактирования)
	EditStatusNotEditing = "Not Editing"
	EditStatusEditing    = "EDIT"
	EditStatusSearch     = "SEARCH"
	EditStatusOutput     = "OUTPUT"
	EditStatusInput      = "INPUT"
	EditStatusCompare    = "COMPARE"
	EditStatusOffset     = "OFFSET"
	EditStatusRestart    = "Restart"
	EditStatusRVRS       = "RVRS"
	EditStatusRTRY       = "RTRY"
	EditStatusRVED       = "RVED"
	EditStatusPTRR       = "PTRR"
	EditStatusAICC       = "AICC"
	EditStatusHPCC       = "HPCC"
	EditStatusNanoHP     = "NANO HP"
	EditStatus5Axis      = "5-AXIS"
	EditStatusWZR        = "WZR"
	EditStatusTCP        = "TCP"
	EditStatusTWP        = "TWP"
	EditStatusTCPAndTWP  = "TCP+TWP"
	EditStatusAPC        = "APC"
	EditStatusProgCheck  = "PRG-CHK"
	EditStatusSTCP       = "S-TCP"
	EditStatusAllSave    = "ALLSAVE"
	EditStatusNotSave    = "NOTSAVE"

	// EditStatus (Специфичные для токарного станка - T-mode)
	EditStatusWorkShift = "Work Shift"
	EditStatusOFSX      = "OFSX"
	EditStatusOFSZ      = "OFSZ"
	EditStatusOFSY      = "OFSY"
	EditStatusTOFS      = "TOFS"

	// EditStatus (Специфичные для фрезерного станка - M-mode)
	EditStatusLabelSkip  = "Label Skip"
	EditStatusHandleMode = "HANDLE"
	EditStatusWorkOffset = "Work Offset"
	EditStatusMemCheck   = "Memory Check"
	EditStatusAIAPC      = "AI APC"
	EditStatusMBLAPC     = "MBL APC"
	EditStatusAIHPCC     = "AI HPCC"
	EditStatusLEN        = "LEN"
	EditStatusRAD        = "RAD"

	// Общее
	StatusUnknown = "UNKNOWN"
)

// ModelUnknownInterpreter предоставляет реализацию по умолчанию для интерпретации состояния станка.
type ModelUnknownInterpreter struct{}

// InterpretMachineState преобразует сырую структуру ODBST в доменную модель UnifiedMachineData.
func (i *ModelUnknownInterpreter) InterpretMachineState(statPtr unsafe.Pointer) *models.UnifiedMachineData {
	stat := (*C.ODBST)(statPtr) // Преобразуем указатель обратно в нужный тип
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
		return TmModeTurning
	case 1:
		return TmModeMilling
	default:
		return StatusUnknown
	}
}

func interpretProgramMode(aut C.short) string {
	switch aut {
	case 0:
		return ProgramModeMDI
	case 1:
		return ProgramModeMemory
	case 2:
		return ProgramModeNoSelection
	case 3:
		return ProgramModeEdit
	case 4:
		return ProgramModeHandle
	case 5:
		return ProgramModeJOG
	case 6:
		return ProgramModeTeachInJOG
	case 7:
		return ProgramModeTeachInHandle
	case 8:
		return ProgramModeIncFeed
	case 9:
		return ProgramModeReference
	case 10:
		return ProgramModeRemote
	default:
		return StatusUnknown
	}
}

func interpretMachineState(run C.short) string {
	switch run {
	case 0:
		return MachineStateReset
	case 1:
		return MachineStateStop
	case 2:
		return MachineStateHold
	case 3:
		return MachineStateStart
	case 4:
		return MachineStateMSTR
	default:
		return StatusUnknown
	}
}

func interpretAxisMovement(motion C.short) string {
	switch motion {
	case 0:
		return AxisMovementNone
	case 1:
		return AxisMovementMotion
	case 2:
		return AxisMovementDwell
	default:
		return StatusUnknown
	}
}

func interpretMstbStatus(mstb C.short) string {
	if mstb == 1 {
		return MstbStatusFIN
	}
	return MstbStatusOther
}

func interpretEmergencyStatus(emergency C.short) string {
	switch emergency {
	case 0:
		return EmergencyStatusNotEmergency
	case 1:
		return EmergencyStatusEmergency
	case 2:
		return EmergencyStatusReset
	case 3:
		return EmergencyStatusWait
	default:
		return StatusUnknown
	}
}

func interpretAlarmStatus(alarm C.short) string {
	switch alarm {
	case 0:
		return AlarmStatusOthers
	case 1:
		return AlarmStatusAlarm
	case 2:
		return AlarmStatusBatteryLow
	case 3:
		return AlarmStatusFan
	case 4:
		return AlarmStatusPSWarning
	case 5:
		return AlarmStatusFSSBWarning
	case 6:
		return AlarmStatusInsulateWarning
	case 7:
		return AlarmStatusEncoderWarning
	case 8:
		return AlarmStatusPMCAlarm
	default:
		return StatusUnknown
	}
}

func interpretEditStatus(tmmode C.short, editValue C.short) string {
	switch tmmode {
	case 0: // T mode (токарный станок)
		switch editValue {
		case 0:
			return EditStatusNotEditing
		case 1:
			return EditStatusEditing
		case 2:
			return EditStatusSearch
		case 3:
			return EditStatusOutput
		case 4:
			return EditStatusInput
		case 5:
			return EditStatusCompare
		case 6:
			return EditStatusOffset
		case 7:
			return EditStatusWorkShift
		case 9:
			return EditStatusRestart
		case 10:
			return EditStatusRVRS
		case 11:
			return EditStatusRTRY
		case 12:
			return EditStatusRVED
		case 14:
			return EditStatusPTRR
		case 16:
			return EditStatusAICC
		case 21:
			return EditStatusHPCC
		case 23:
			return EditStatusNanoHP
		case 25:
			return EditStatus5Axis
		case 26:
			return EditStatusOFSX
		case 27:
			return EditStatusOFSZ
		case 28:
			return EditStatusWZR
		case 29:
			return EditStatusOFSY
		case 31:
			return EditStatusTOFS
		case 39:
			return EditStatusTCP
		case 40:
			return EditStatusTWP
		case 41:
			return EditStatusTCPAndTWP
		case 42, 44:
			return EditStatusAPC
		case 43:
			return EditStatusProgCheck
		case 45:
			return EditStatusSTCP
		case 59:
			return EditStatusAllSave
		case 60:
			return EditStatusNotSave
		default:
			return StatusUnknown
		}
	case 1: // M mode (фрезерный станок)
		switch editValue {
		case 0:
			return EditStatusNotEditing
		case 1:
			return EditStatusEditing
		case 2:
			return EditStatusSearch
		case 3:
			return EditStatusOutput
		case 4:
			return EditStatusInput
		case 5:
			return EditStatusCompare
		case 6:
			return EditStatusLabelSkip
		case 7:
			return EditStatusRestart
		case 8:
			return EditStatusHPCC
		case 9:
			return EditStatusPTRR
		case 10:
			return EditStatusRVRS
		case 11:
			return EditStatusRTRY
		case 12:
			return EditStatusRVED
		case 13:
			return EditStatusHandleMode
		case 14:
			return EditStatusOffset
		case 15:
			return EditStatusWorkOffset
		case 16:
			return EditStatusAICC
		case 17:
			return EditStatusMemCheck
		case 21:
			return EditStatusAIAPC
		case 22:
			return EditStatusMBLAPC
		case 23:
			return EditStatusNanoHP
		case 24:
			return EditStatusAIHPCC
		case 25:
			return EditStatus5Axis
		case 26:
			return EditStatusLEN
		case 27:
			return EditStatusRAD
		case 28:
			return EditStatusWZR
		case 39:
			return EditStatusTCP
		case 40:
			return EditStatusTWP
		case 41:
			return EditStatusTCPAndTWP
		case 42, 44:
			return EditStatusAPC
		case 43:
			return EditStatusProgCheck
		case 45:
			return EditStatusSTCP
		case 59:
			return EditStatusAllSave
		case 60:
			return EditStatusNotSave
		default:
			return StatusUnknown
		}
	default:
		return StatusUnknown
	}
}
