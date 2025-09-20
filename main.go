package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/iwtcode/fanucService/internal/config"
	"github.com/iwtcode/fanucService/internal/domain"
	"github.com/iwtcode/fanucService/internal/focas"
	models "github.com/iwtcode/fanucService/pkg/focas/models"
	"github.com/joho/godotenv"
)

// runStep - обертка, которая теперь управляет полным циклом операции:
// 1. Подключается к ЧПУ.
// 2. Выполняет переданную функцию (fn).
// 3. Отключается от ЧПУ.
func runStep(name string, cfg *config.Config, fn func(handle uint16) error) {
	log.Printf("--- Запуск шага: %s ---", name)

	// 1. Подключение перед каждым шагом
	h, err := focas.Connect(cfg.IP, cfg.Port, cfg.TimeoutMs)
	if err != nil {
		log.Fatalf("Ошибка подключения на шаге %s: %v", name, err)
	}
	// 3. Гарантированное отключение после завершения шага
	defer focas.Disconnect(h)

	// 2. Выполнение основной логики
	if err := fn(h); err != nil {
		log.Fatalf("Ошибка выполнения на шаге %s: %v", name, err)
	}

	log.Printf("--- Шаг %s выполнен успешно ---", name)
	fmt.Println("==================================================")
}

func main() {
	// 1) Загрузка конфигурации
	err := godotenv.Load("./.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file. Using default values or environment variables: %v", err)
	}

	cfg := config.Load()
	log.Printf("Конфигурация загружена: IP=%s, Port=%d, Timeout=%dms", cfg.IP, cfg.Port, cfg.TimeoutMs)

	// 2) Инициализация FOCAS (делается один раз при старте приложения)
	if err := focas.Startup(3, cfg.LogPath); err != nil {
		log.Fatalf("FOCAS startup error: %v", err)
	}
	log.Println("FOCAS успешно инициализирован.")

	// Глобальное подключение больше не нужно. Логика переехала в runStep.

	// 4) Чтение системной информации
	var internalSystemInfo *domain.SystemInfo
	runStep("ReadSystemInfo", cfg, func(h uint16) error {
		var err error
		internalSystemInfo, err = focas.ReadSystemInfo(h)
		if err != nil {
			return err
		}
		pkgSystemInfo := models.SystemInfo{
			Manufacturer:   internalSystemInfo.Manufacturer,
			Model:          internalSystemInfo.Model,
			Series:         internalSystemInfo.Series,
			Version:        internalSystemInfo.Version,
			ControlledAxes: internalSystemInfo.ControlledAxes,
		}
		printAsJSON("SystemInfo", pkgSystemInfo)
		return nil
	})

	// 5) Чтение информации о программе
	runStep("ReadProgramInfo", cfg, func(h uint16) error {
		internalProgInfo, err := focas.ReadProgram(h)
		if err != nil {
			log.Printf("Предупреждение: Не удалось прочитать информацию о программе: %v", err)
			return nil // Не считаем это фатальной ошибкой
		}
		pkgProgInfo := models.ProgramInfo{
			Name:      internalProgInfo.Name,
			Number:    internalProgInfo.Number,
			GCodeLine: internalProgInfo.CurrentGCode,
		}
		printAsJSON("ProgramInfo", pkgProgInfo)
		return nil
	})

	// 6) Чтение информации об осях
	runStep("ReadAxisData", cfg, func(h uint16) error {
		if internalSystemInfo == nil {
			return fmt.Errorf("системная информация не была получена на предыдущем шаге, невозможно прочитать данные осей")
		}
		internalAxisInfos, err := focas.ReadAxisData(h, internalSystemInfo.ControlledAxes, internalSystemInfo.MaxAxis)
		if err != nil {
			return err
		}
		pkgAxisInfos := make([]models.AxisInfo, len(internalAxisInfos))
		for i, axis := range internalAxisInfos {
			pkgAxisInfos[i] = models.AxisInfo{
				Name: axis.Name, Position: axis.Position, LoadPercent: axis.LoadPercent,
				ServoTemperature: axis.ServoTemperature, CoderTemperature: axis.CoderTemperature,
				PowerConsumption: axis.PowerConsumption, Diag301: axis.Diag301,
			}
		}
		printAsJSON("AxisData", pkgAxisInfos)
		return nil
	})

	// 7) Чтение информации о шпинделях
	runStep("ReadSpindleData", cfg, func(h uint16) error {
		internalSpindleInfos, err := focas.ReadSpindleData(h)
		if err != nil {
			return err
		}
		pkgSpindleInfos := make([]models.SpindleInfo, len(internalSpindleInfos))
		for i, spindle := range internalSpindleInfos {
			pkgSpindleInfos[i] = models.SpindleInfo{Number: spindle.Number, SpeedRPM: spindle.SpeedRPM}
		}
		printAsJSON("SpindleData", pkgSpindleInfos)
		return nil
	})

	// 8) Чтение полного состояния станка
	runStep("ReadMachineState", cfg, func(h uint16) error {
		internalState, err := focas.ReadMachineState(h)
		if err != nil {
			return err
		}
		pkgState := models.MachineState{
			TmMode: internalState.TmMode, ProgramMode: internalState.ProgramMode, MachineState: internalState.MachineState,
			AxisMovementStatus: internalState.AxisMovementStatus, MstbStatus: internalState.MstbStatus,
			EmergencyStatus: internalState.EmergencyStatus, AlarmStatus: internalState.AlarmStatus, EditStatus: internalState.EditStatus,
		}
		printAsJSON("MachineState", pkgState)
		return nil
	})

	// 9) Чтение полного текста программы
	runStep("GetControlProgram", cfg, func(h uint16) error {
		gcode, err := focas.GetControlProgram(h)
		if err != nil {
			log.Printf("Предупреждение: Не удалось прочитать содержимое программы: %v", err)
			return nil // Не считаем это фатальной ошибкой
		}
		filePath := "g_code_from_main.log"
		err = os.WriteFile(filePath, []byte(gcode), 0644)
		if err != nil {
			return fmt.Errorf("не удалось записать G-код в файл %s: %w", filePath, err)
		}
		log.Printf("G-код программы успешно сохранен в %s", filePath)
		return nil
	})

	log.Println("Сбор данных завершен.")
}

// printAsJSON форматирует данные в JSON и выводит в лог
func printAsJSON(name string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Ошибка маршалинга JSON для %s: %v", name, err)
		return
	}
	fmt.Printf("--- %s ---\n%s\n", name, string(jsonData))
}
