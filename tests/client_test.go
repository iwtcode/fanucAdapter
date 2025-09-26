package tests

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	fanuc "github.com/iwtcode/fanucService"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// setupTest загружает конфигурацию и создает новый клиент
func setupTest(t *testing.T) *fanuc.Client {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file from ../.env. Using default values or environment variables: %v", err)
	}

	cfg := fanuc.Load() // Используем fanuc.Load()
	log.Printf("Конфигурация загружена: IP=%s, Port=%d", cfg.IP, cfg.Port)
	require.NotNil(t, cfg, "Конфигурация не была загружена")

	log.Printf("Подключение к %s:%d ...", cfg.IP, cfg.Port)
	c, err := fanuc.New(cfg) // Используем fanuc.New()
	require.NoError(t, err, "Не удалось создать FOCAS клиент")
	require.NotNil(t, c, "Клиент не должен быть nil")
	log.Println("Успешно подключено!")

	return c
}

// logAsJSON форматирует данные в JSON и выводит в лог теста
func logAsJSON(t *testing.T, name string, data interface{}) {
	t.Helper()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err, "Ошибка маршалинга JSON для %s", name)
	log.Printf("--- %s ---\n%s", name, string(jsonData))
}

func TestReadSystemInfo(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	sysInfo := c.GetSystemInfo()
	require.NotNil(t, sysInfo)
	logAsJSON(t, "SystemInfo", sysInfo)
}

func TestReadProgramInfo(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	progInfo, err := c.GetProgramInfo()
	if err != nil {
		t.Skipf("Пропускаем тест информации о программе: %v", err)
	}

	logAsJSON(t, "ProgramInfo", progInfo)
}

func TestReadAxisData(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	axisInfos, err := c.GetAxisData()
	require.NoError(t, err, "Не удалось прочитать информацию об осях")

	logAsJSON(t, "AxisData", axisInfos)
}

func TestReadSpindleData(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	spindleInfos, err := c.GetSpindleData()
	require.NoError(t, err, "Не удалось прочитать информацию о шпинделях")

	logAsJSON(t, "SpindleData", spindleInfos)
}

func TestReadMachineState(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	state, err := c.GetMachineState()
	require.NoError(t, err, "Не удалось прочитать состояние станка")

	logAsJSON(t, "MachineState", state)
}

func TestGetControlProgram(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	gcode, err := c.GetControlProgram()
	if err != nil {
		t.Skipf("Пропускаем тест содержимого программы: %v", err)
	}

	filePath := "g_code.log"
	err = os.WriteFile(filePath, []byte(gcode), 0644)
	require.NoError(t, err, "Не удалось записать G-код в файл %s", filePath)

	log.Printf("G-код программы успешно сохранен в %s", filePath)
}
