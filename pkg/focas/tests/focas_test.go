package tests

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/iwtcode/fanucService/internal/config"
	"github.com/iwtcode/fanucService/internal/focas"
	models "github.com/iwtcode/fanucService/pkg/focas/models"
	"github.com/joho/godotenv"
)

// TestFanucConnectionAndDataReading - основной тест для проверки всех функций
func TestFanucConnectionAndDataReading(t *testing.T) {
	// 1) Загрузка конфигурации
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file. Using default values or environment variables: %v", err)
	}

	cfg := config.Load()
	t.Logf("Конфигурация загружена: IP=%s, Port=%d", cfg.IP, cfg.Port)

	// 2) Инициализация FOCAS
	if err := focas.Startup(3, cfg.LogPath); err != nil {
		t.Fatalf("FOCAS startup error: %v", err)
	}
	t.Log("FOCAS успешно инициализирован.")

	// 3) Подключение к ЧПУ
	t.Logf("Подключение к %s:%d ...", cfg.IP, cfg.Port)
	h, err := focas.Connect(cfg.IP, cfg.Port, cfg.TimeoutMs)
	if err != nil {
		t.Fatalf("Не удалось подключиться: %v", err)
	}
	defer focas.Disconnect(h)
	t.Log("Успешно подключено!")

	// 4) Чтение системной информации
	t.Run("ReadSystemInfo", func(t *testing.T) {
		internalSystemInfo, err := focas.ReadSystemInfo(h)
		if err != nil {
			t.Errorf("Не удалось прочитать системную информацию: %v", err)
			return
		}
		pkgSystemInfo := models.SystemInfo{
			Manufacturer:   internalSystemInfo.Manufacturer,
			Model:          internalSystemInfo.Model,
			Series:         internalSystemInfo.Series,
			Version:        internalSystemInfo.Version,
			ControlledAxes: internalSystemInfo.ControlledAxes,
		}
		logAsJSON(t, "SystemInfo", pkgSystemInfo)
	})

	// 5) Чтение информации о программе
	t.Run("ReadProgramInfo", func(t *testing.T) {
		internalProgInfo, err := focas.ReadProgram(h)
		if err != nil {
			log.Printf("Не удалось прочитать информацию о программе: %v", err)
			t.Skipf("Пропускаем тест информации о программе: %v", err)
		} else {
			pkgProgInfo := models.ProgramInfo{
				Name:      internalProgInfo.Name,
				Number:    internalProgInfo.Number,
				GCodeLine: internalProgInfo.CurrentGCode,
			}
			logAsJSON(t, "ProgramInfo", pkgProgInfo)
		}
	})

	// 6) Чтение информации об осях
	t.Run("ReadAxisData", func(t *testing.T) {
		sysInfo, err := focas.ReadSystemInfo(h)
		if err != nil {
			t.Fatalf("Не удалось получить SystemInfo для чтения осей: %v", err)
		}

		internalAxisInfos, err := focas.ReadAxisData(h, sysInfo.ControlledAxes, sysInfo.MaxAxis)
		if err != nil {
			t.Errorf("Не удалось прочитать информацию об осях: %v", err)
			return
		}

		pkgAxisInfos := make([]models.AxisInfo, len(internalAxisInfos))
		for i, axis := range internalAxisInfos {
			pkgAxisInfos[i] = models.AxisInfo{
				Name:             axis.Name,
				Position:         axis.Position,
				LoadPercent:      axis.LoadPercent,
				ServoTemperature: axis.ServoTemperature,
				CoderTemperature: axis.CoderTemperature,
				PowerConsumption: axis.PowerConsumption,
				Diag301:          axis.Diag301,
			}
		}
		logAsJSON(t, "AxisData", pkgAxisInfos)
	})

	// 7) Чтение полного состояния станка
	t.Run("ReadMachineState", func(t *testing.T) {
		internalState, err := focas.ReadMachineState(h)
		if err != nil {
			t.Errorf("Не удалось прочитать состояние станка: %v", err)
			return
		}
		pkgState := models.MachineState{
			TmMode:             internalState.TmMode,
			ProgramMode:        internalState.ProgramMode,
			MachineState:       internalState.MachineState,
			AxisMovementStatus: internalState.AxisMovementStatus,
			MstbStatus:         internalState.MstbStatus,
			EmergencyStatus:    internalState.EmergencyStatus,
			AlarmStatus:        internalState.AlarmStatus,
			EditStatus:         internalState.EditStatus,
		}
		logAsJSON(t, "MachineState", pkgState)
	})

	// 8) Чтение полного текста программы
	t.Run("GetControlProgram", func(t *testing.T) {
		gcode, err := focas.GetControlProgram(h)
		if err != nil {
			log.Printf("Не удалось прочитать содержимое программы: %v", err)
			t.Skipf("Пропускаем тест содержимого программы: %v", err)
		} else {
			filePath := "g_code.log"
			err := os.WriteFile(filePath, []byte(gcode), 0644)
			if err != nil {
				t.Errorf("Не удалось записать G-код в файл %s: %v", filePath, err)
			} else {
				t.Logf("G-код программы успешно сохранен в %s", filePath)
			}
		}
	})
}

// logAsJSON форматирует данные в JSON и выводит в лог теста
func logAsJSON(t *testing.T, name string, data interface{}) {
	t.Helper()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("Ошибка маршалинга JSON для %s: %v", name, err)
		return
	}
	t.Logf("--- %s ---\n%s", name, string(jsonData))
}
