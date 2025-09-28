package focas

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iwtcode/fanucService/models"
)

// AggregateAllData собирает все доступные данные со станка асинхронно.
func (a *FocasAdapter) AggregateAllData() (*models.AggregatedData, error) {
	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex // для безопасной записи ошибок из горутин

	// Переменные для хранения результатов
	var machineState *models.UnifiedMachineData
	var axisData []models.AxisInfo
	var spindleData []models.SpindleInfo
	var programInfo *models.ProgramInfo

	// 1. Получение состояния станка
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		machineState, err = a.ReadMachineState()
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to read machine state: %w", err))
			mu.Unlock()
		}
	}()

	// 2. Получение данных по осям
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		axisData, err = a.ReadAxisData()
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to read axis data: %w", err))
			mu.Unlock()
		}
	}()

	// 3. Получение данных по шпинделям
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		spindleData, err = a.ReadSpindleData()
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to read spindle data: %w", err))
			mu.Unlock()
		}
	}()

	// 4. Получение информации о программе
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		programInfo, err = a.ReadProgram()
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to read program info: %w", err))
			mu.Unlock()
		}
	}()

	// Ожидаем завершения всех горутин
	wg.Wait()

	// Проверяем, были ли ошибки
	if len(errs) > 0 {
		var errorStrings []string
		for _, e := range errs {
			errorStrings = append(errorStrings, e.Error())
		}
		return nil, fmt.Errorf("aggregation failed with multiple errors: %s", strings.Join(errorStrings, "; "))
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
	}

	return data, nil
}
