package models

import "time"

// ProgramInfo содержит информацию о выполняемой программе.
type ProgramInfo struct {
	Name   string `json:"name"`
	Number int64  `json:"number"`
}

// SystemInfo содержит системную информацию о станке.
type SystemInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Series       string `json:"series"`
	Version      string `json:"version"`
}

// UnifiedMachineData содержит полное унифицированное состояние станка.
type UnifiedMachineData struct {
	TmMode             string `json:"tm_mode"`
	ProgramMode        string `json:"program_mode"`
	MachineState       string `json:"machine_state"`
	AxisMovementStatus string `json:"axis_movement_status"`
	MstbStatus         string `json:"mstb_status"`
	EmergencyStatus    string `json:"emergency_status"`
	AlarmStatus        string `json:"alarm_status"`
	EditStatus         string `json:"edit_status"`
}

// FullMachineData - это агрегированная структура для отправки в Kafka.
type FullMachineData struct {
	SessionID   string              `json:"session_id"`
	Endpoint    string              `json:"endpoint"`
	Timestamp   time.Time           `json:"timestamp"`
	SystemInfo  *SystemInfo         `json:"system_info"`
	ProgramInfo *ProgramInfo        `json:"program_info"`
	StatusInfo  *UnifiedMachineData `json:"status_info"`
}
