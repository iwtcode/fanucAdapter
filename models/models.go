package models

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
	MaxAxes        int16  `json:"max_axes"` // Переименовано
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
	Alarms             []AlarmDetail `json:"alarms"` // Добавлено поле для ошибок
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
