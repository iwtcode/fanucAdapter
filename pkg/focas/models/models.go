package focas

// ProgramInfo представляет информацию о программе на уровне пакета
type ProgramInfo struct {
	Name   string `json:"name"`
	Number int64  `json:"number"`
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
	Name     string  `json:"name"`
	Position float64 `json:"position"`
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
