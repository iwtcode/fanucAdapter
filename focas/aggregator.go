package focas

import (
	"fmt"
	"log"
	"time"

	"github.com/iwtcode/fanucAdapter/models"
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

	// 6. Получение данных о контурной подаче
	contourFeedRate, err := a.ReadContourFeedRate()
	if err != nil {
		return nil, fmt.Errorf("failed to read contour feed rate: %w", err)
	}

	// 7. Получение данных о коррекции JOG
	jogOverride, err := a.ReadJogOverride()
	if err != nil {
		log.Printf("Warning: failed to read jog override: %v", err)
		jogOverride = 0 // Устанавливаем значение по умолчанию в случае ошибки
	}

	// 8. Получение параметров (счетчики, время)
	paramInfo, err := a.ReadParameterInfo()
	if err != nil {
		// Ошибки уже логируются внутри ReadParameterInfo, поэтому здесь просто продолжаем
		log.Printf("Warning: one or more parameters could not be read: %v", err)
		// Инициализируем пустой структурой, чтобы избежать nil pointer dereference
		paramInfo = &models.ParameterInfo{}
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
		ContourFeedRate:    contourFeedRate,
		ActualFeedRate:     feedInfo.ActualFeedRate,
		FeedOverride:       feedInfo.FeedOverride,
		JogOverride:        jogOverride,
		PartsCount:         paramInfo.PartsCount,
		PowerOnTime:        paramInfo.PowerOnTime,
		OperatingTime:      paramInfo.OperatingTime,
		CycleTime:          paramInfo.CycleTime,
		CuttingTime:        paramInfo.CuttingTime,
	}

	return data, nil
}
