package domain

// ProgramInfo содержит информацию о выполняемой программе
type ProgramInfo struct {
	Name         string
	Number       int64
	CurrentGCode string
}

// ControlProgram содержит информацию о выполняемой программе и ее содержимое
type ControlProgram struct {
	ProgramInfo
	GCode string
}

// SystemInfo содержит системную информацию о станке
type SystemInfo struct {
	Manufacturer   string
	Model          string
	Series         string
	Version        string
	MaxAxis        int16
	ControlledAxes int16
}

// AxisInfo содержит информацию об оси
type AxisInfo struct {
	Name     string
	Position float64
}

// UnifiedMachineData содержит полное унифицированное состояние станка
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
