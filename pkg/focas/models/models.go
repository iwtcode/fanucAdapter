package focas

// ProgramInfo представляет информацию о программе на уровне пакета
type ProgramInfo struct {
	Name      string `json:"program_name"`
	Number    int64  `json:"program_number"`
	GCodeLine string `json:"g_code_line"`
}

// ControlProgram представляет полные данные о программе для вывода в API/тестах
type ControlProgram struct {
	Name      string `json:"program_name"`
	Number    int64  `json:"program_number"`
	GCodeLine string `json:"g_code_line"`
	GCode     string `json:"g_code"`
}

// SystemInfo представляет системную информацию на уровне пакета
type SystemInfo struct {
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	Series         string `json:"series"`
	Version        string `json:"version"`
	ControlledAxes int16  `json:"controlled_axes"`
}

// AxisInfo представляет информацию об оси на уровне пакета
type AxisInfo struct {
	Name             string  `json:"name"`
	Position         float64 `json:"position"`
	LoadPercent      float64 `json:"load_percent"`
	ServoTemperature int32   `json:"servo_temperature"`
	CoderTemperature int32   `json:"coder_temperature"`
	PowerConsumption int32   `json:"power_consumption"`
	Diag301          float64 `json:"diag_301"`
}

// SpindleInfo представляет информацию о шпинделе на уровне пакета
type SpindleInfo struct {
	Number           int16   `json:"number"`
	SpeedRPM         int32   `json:"speed_rpm"`
	LoadPercent      float64 `json:"load_percent"`
	OverridePercent  int16   `json:"override_percent"`
	PowerConsumption int32   `json:"power_consumption"`
	Diag411Value     int32   `json:"diag_411_value"`
}

// MachineState представляет состояние станка на уровне пакета
type MachineState struct {
	TmMode             string `json:"tm_mode"`
	ProgramMode        string `json:"program_mode"`
	MachineState       string `json:"machine_state"`
	AxisMovementStatus string `json:"axis_movement_status"`
	MstbStatus         string `json:"mstb_status"`
	EmergencyStatus    string `json:"emergency_status"`
	AlarmStatus        string `json:"alarm_status"`
	EditStatus         string `json:"edit_status"`
}
