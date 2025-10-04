package models

import "time"

// ProgramInfo содержит информацию о выполняемой программе
type ProgramInfo struct {
	Name         string `json:"name"`
	Number       int64  `json:"number"`
	CurrentGCode string `json:"current_g_code"`
}

// ControlProgram содержит информацию о выполняемой программе и ее содержимое
type ControlProgram struct {
	ProgramInfo
	GCode string `json:"g_code"`
}

// SystemInfo содержит системную информацию о станке
type SystemInfo struct {
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	Series         string `json:"series"`
	Version        string `json:"version"`
	MaxAxes        int16  `json:"max_axes"`
	ControlledAxes int16  `json:"controlled_axes"`
}

// AxisInfo содержит информацию об оси
type AxisInfo struct {
	Name             string  `json:"name"`
	Position         float64 `json:"position"`
	LoadPercent      float64 `json:"load_percent"`
	ServoTemperature int32   `json:"servo_temperature"`
	CoderTemperature int32   `json:"coder_temperature"`
	PowerConsumption int32   `json:"power_consumption"`
	Diag301          float64 `json:"diag_301"`
}

// AlarmDetail содержит детальную информацию об одной ошибке
type AlarmDetail struct {
	ErrorCode            string `json:"error_code"`
	ErrorTypeDescription string `json:"error_type_description"`
	ErrorMessage         string `json:"error_message"`
}

// UnifiedMachineData содержит полное унифицированное состояние станка
type UnifiedMachineData struct {
	TmMode             string        `json:"tm_mode"`
	ProgramMode        string        `json:"program_mode"`
	MachineState       string        `json:"machine_state"`
	AxisMovementStatus string        `json:"axis_movement_status"`
	MstbStatus         string        `json:"mstb_status"`
	EmergencyStatus    string        `json:"emergency_status"`
	AlarmStatus        string        `json:"alarm_status"`
	EditStatus         string        `json:"edit_status"`
	Alarms             []AlarmDetail `json:"alarms"`
}

// SpindleInfo содержит информацию о шпинделе
type SpindleInfo struct {
	Number           int16   `json:"number"`
	SpeedRPM         int32   `json:"speed_rpm"`
	LoadPercent      float64 `json:"load_percent"`
	OverridePercent  int16   `json:"override_percent"`
	PowerConsumption int32   `json:"power_consumption"`
	Diag411Value     int32   `json:"diag_411_value"`
}

// CurrentProgramInfo содержит упрощенную информацию о текущей программе для AggregatedData.
type CurrentProgramInfo struct {
	ProgramName   string `json:"program_name"`
	ProgramNumber int64  `json:"program_number"`
	GCodeLine     string `json:"g_code_line"`
}

// FeedInfo содержит информацию о скорости подачи и коррекции.
type FeedInfo struct {
	ActualFeedRate int32 `json:"actual_feed_rate"`
	FeedOverride   int16 `json:"feed_override"`
}

// AggregatedData содержит полную сводку данных о станке.
type AggregatedData struct {
	MachineID          string             `json:"machine_id"`
	Timestamp          time.Time          `json:"timestamp"`
	IsEnabled          bool               `json:"is_enabled"`
	IsEmergency        bool               `json:"is_emergency"`
	MachineState       string             `json:"machine_state"`
	ProgramMode        string             `json:"program_mode"`
	TmMode             string             `json:"tm_mode"`
	AxisMovementStatus string             `json:"axis_movement_status"`
	MstbStatus         string             `json:"mstb_status"`
	EmergencyStatus    string             `json:"emergency_status"`
	AlarmStatus        string             `json:"alarm_status"`
	EditStatus         string             `json:"edit_status"`
	AxisInfos          []AxisInfo         `json:"axis_infos"`
	HasAlarms          bool               `json:"has_alarms"`
	Alarms             []AlarmDetail      `json:"alarms"`
	CurrentProgram     CurrentProgramInfo `json:"current_program"`
	SpindleInfos       []SpindleInfo      `json:"spindle_infos"`
	ContourFeedRate    int32              `json:"contour_feed_rate"`
	ActualFeedRate     int32              `json:"actual_feed_rate"`
	FeedOverride       int16              `json:"feed_override"`
}
