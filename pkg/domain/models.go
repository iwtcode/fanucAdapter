package domain

// ProgramInfo содержит информацию о выполняемой программе.
type ProgramInfo struct {
	Name   string
	Number int64
}

// SystemInfo содержит системную информацию о станке.
type SystemInfo struct {
	Manufacturer string
	Model        string
	Series       string
	Version      string
}

// UnifiedMachineData содержит полное унифицированное состояние станка.
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
