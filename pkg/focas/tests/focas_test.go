package tests

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/iwtcode/fanucService/internal/config"
	"github.com/iwtcode/fanucService/internal/focas"
	models "github.com/iwtcode/fanucService/pkg/focas/models"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var (
	testConfig *config.Config
	once       sync.Once
)

// setupTest загружает конфигурацию, инициализирует FOCAS и подключается к станку
func setupTest(t *testing.T) (uint16, *config.Config) {
	once.Do(func() {
		err := godotenv.Load("../../../.env")
		if err != nil {
			log.Printf("Warning: Could not load .env file. Using default values or environment variables: %v", err)
		}

		cfg := config.Load()
		log.Printf("Конфигурация загружена: IP=%s, Port=%d", cfg.IP, cfg.Port)
		testConfig = cfg

		if err := focas.Startup(3, cfg.LogPath); err != nil {
			log.Fatalf("FOCAS startup error: %v", err)
		}
		log.Println("FOCAS успешно инициализирован.")
	})

	require.NotNil(t, testConfig, "Конфигурация не была загружена")

	log.Printf("Подключение к %s:%d ...", testConfig.IP, testConfig.Port)
	h, err := focas.Connect(testConfig.IP, testConfig.Port, testConfig.TimeoutMs)
	require.NoError(t, err, "Не удалось подключиться")
	require.NotZero(t, h, "Получен неверный хендл")
	log.Println("Успешно подключено!")

	return h, testConfig
}

// logAsJSON форматирует данные в JSON и выводит в лог теста
func logAsJSON(t *testing.T, name string, data interface{}) {
	t.Helper()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err, "Ошибка маршалинга JSON для %s", name)
	t.Logf("--- %s ---\n%s", name, string(jsonData))
}

func TestReadSystemInfo(t *testing.T) {
	h, _ := setupTest(t)
	defer focas.Disconnect(h)

	internalSystemInfo, err := focas.ReadSystemInfo(h)
	require.NoError(t, err, "Не удалось прочитать системную информацию")

	pkgSystemInfo := models.SystemInfo{
		Manufacturer:   internalSystemInfo.Manufacturer,
		Model:          internalSystemInfo.Model,
		Series:         internalSystemInfo.Series,
		Version:        internalSystemInfo.Version,
		ControlledAxes: internalSystemInfo.ControlledAxes,
	}
	logAsJSON(t, "SystemInfo", pkgSystemInfo)
}

func TestReadProgramInfo(t *testing.T) {
	h, _ := setupTest(t)
	defer focas.Disconnect(h)

	internalProgInfo, err := focas.ReadProgram(h)
	if err != nil {
		t.Skipf("Пропускаем тест информации о программе: %v", err)
	}

	pkgProgInfo := models.ProgramInfo{
		Name:      internalProgInfo.Name,
		Number:    internalProgInfo.Number,
		GCodeLine: internalProgInfo.CurrentGCode,
	}
	logAsJSON(t, "ProgramInfo", pkgProgInfo)
}

func TestReadAxisData(t *testing.T) {
	h, _ := setupTest(t)
	defer focas.Disconnect(h)

	sysInfo, err := focas.ReadSystemInfo(h)
	require.NoError(t, err, "Не удалось получить SystemInfo для чтения осей")

	internalAxisInfos, err := focas.ReadAxisData(h, sysInfo.ControlledAxes, sysInfo.MaxAxis)
	require.NoError(t, err, "Не удалось прочитать информацию об осях")

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
}

func TestReadSpindleData(t *testing.T) {
	h, _ := setupTest(t)
	defer focas.Disconnect(h)

	internalSpindleInfos, err := focas.ReadSpindleData(h)
	require.NoError(t, err, "Не удалось прочитать информацию о шпинделях")

	pkgSpindleInfos := make([]models.SpindleInfo, len(internalSpindleInfos))
	for i, spindle := range internalSpindleInfos {
		pkgSpindleInfos[i] = models.SpindleInfo{
			Number:           spindle.Number,
			SpeedRPM:         spindle.SpeedRPM,
			LoadPercent:      spindle.LoadPercent,
			OverridePercent:  spindle.OverridePercent,
			PowerConsumption: spindle.PowerConsumption,
		}
	}
	logAsJSON(t, "SpindleData", pkgSpindleInfos)
}

func TestReadMachineState(t *testing.T) {
	h, _ := setupTest(t)
	defer focas.Disconnect(h)

	internalState, err := focas.ReadMachineState(h)
	require.NoError(t, err, "Не удалось прочитать состояние станка")

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
}

func TestGetControlProgram(t *testing.T) {
	h, _ := setupTest(t)
	defer focas.Disconnect(h)

	gcode, err := focas.GetControlProgram(h)
	if err != nil {
		t.Skipf("Пропускаем тест содержимого программы: %v", err)
	}

	filePath := "g_code.log"
	err = os.WriteFile(filePath, []byte(gcode), 0644)
	require.NoError(t, err, "Не удалось записать G-код в файл %s", filePath)

	t.Logf("G-код программы успешно сохранен в %s", filePath)
}
