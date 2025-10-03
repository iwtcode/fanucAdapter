package focas

import (
	"fmt"
	"time"

	"github.com/iwtcode/fanucService/models"
)

// AggregateAllData собирает все доступные данные со станка последовательно.
func (a *FocasAdapter) AggregateAllData() (*models.AggregatedData, error) {
	// 1. Получение состояния станка
	machineState, err := a.ReadMachineState()
	if err != nil {
		return nil, fmt.Errorf("failed to read machine state: %w", err)
	}

	// 2. Получение данных по осям
	axisData, err := a.ReadAxisData()
	if err != nil {
		return nil, fmt.Errorf("failed to read axis data: %w", err)
	}

	// 3. Получение данных по шпинделям
	spindleData, err := a.ReadSpindleData()
	if err != nil {
		return nil, fmt.Errorf("failed to read spindle data: %w", err)
	}

	// 4. Получение информации о программе
	programInfo, err := a.ReadProgram()
	if err != nil {
		return nil, fmt.Errorf("failed to read program info: %w", err)
	}

	// 5. Получение данных о подаче
	feedInfo, err := a.ReadFeedData()
	if err != nil {
		return nil, fmt.Errorf("failed to read feed data: %w", err)
	}

	// Сборка финальной структуры
	isEmergency := machineState.EmergencyStatus != "Not Emergency"
	hasAlarms := len(machineState.Alarms) > 0

	currentProg := models.CurrentProgramInfo{}
	if programInfo != nil {
		currentProg.ProgramName = programInfo.Name
		currentProg.ProgramNumber = programInfo.Number
		currentProg.GCodeLine = programInfo.CurrentGCode
	}

	data := &models.AggregatedData{
		MachineID:          a.ip,
		Timestamp:          time.Now().UTC(),
		IsEnabled:          true,
		IsEmergency:        isEmergency,
		MachineState:       machineState.MachineState,
		ProgramMode:        machineState.ProgramMode,
		TmMode:             machineState.TmMode,
		AxisMovementStatus: machineState.AxisMovementStatus,
		MstbStatus:         machineState.MstbStatus,
		EmergencyStatus:    machineState.EmergencyStatus,
		AlarmStatus:        machineState.AlarmStatus,
		EditStatus:         machineState.EditStatus,
		HasAlarms:          hasAlarms,
		Alarms:             machineState.Alarms,
		AxisInfos:          axisData,
		SpindleInfos:       spindleData,
		CurrentProgram:     currentProg,
		FeedInfo:           feedInfo,
	}

	return data, nil
}
