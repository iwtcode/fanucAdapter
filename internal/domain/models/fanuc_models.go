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

// CurrentProgramInfoKafka содержит информацию о текущей выполняемой программе для Kafka.
type CurrentProgramInfoKafka struct {
	Program        string `json:"program,omitempty"`
	ProgramComment string `json:"program_comment,omitempty"`
}

// MachineDataKafka - это агрегированная структура для отправки в Kafka, совместимая с MTConnect.
type MachineDataKafka struct {
	MachineId           string                   `json:"machine_id"`
	Id                  string                   `json:"id,omitempty"`
	Timestamp           string                   `json:"timestamp"`
	IsEnabled           interface{}              `json:"is_enabled,omitempty"`
	IsInEmergency       interface{}              `json:"is_in_emergency,omitempty"`
	MachineState        string                   `json:"machine_state"`
	ProgramMode         string                   `json:"program_mode"`
	TmMode              string                   `json:"tm_mode"`
	HandleRetraceStatus interface{}              `json:"handle_retrace_status,omitempty"`
	AxisMovementStatus  interface{}              `json:"axis_movement_status"`
	MstbStatus          string                   `json:"mstb_status"`
	EmergencyStatus     string                   `json:"emergency_status"`
	AlarmStatus         string                   `json:"alarm_status"`
	EditStatus          string                   `json:"edit_status"`
	ManualMode          interface{}              `json:"manual_mode,omitempty"`
	WriteStatus         string                   `json:"write_status,omitempty"`
	LabelSkipStatus     interface{}              `json:"label_skip_status,omitempty"`
	WarningStatus       string                   `json:"warning_status,omitempty"`
	BatteryStatus       interface{}              `json:"battery_status,omitempty"`
	ActiveToolNumber    string                   `json:"active_tool_number,omitempty"`
	ToolOffsetNumber    string                   `json:"tool_offset_number,omitempty"`
	Alarms              []map[string]interface{} `json:"alarms,omitempty"`
	HasAlarms           interface{}              `json:"has_alarms,omitempty"`
	PartsCount          map[string]string        `json:"parts_count,omitempty"`
	AccumulatedTime     map[string]string        `json:"accumulated_time,omitempty"`
	CurrentProgram      *CurrentProgramInfoKafka `json:"current_program,omitempty"`
}
