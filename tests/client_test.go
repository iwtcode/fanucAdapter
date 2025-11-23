package tests

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	fanuc "github.com/iwtcode/fanucAdapter"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) *fanuc.Client {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file from ../.env. Using default values or environment variables: %v", err)
	}

	cfg := fanuc.Load()
	log.Printf("Конфигурация загружена: IP=%s, Port=%d", cfg.IP, cfg.Port)
	require.NotNil(t, cfg, "Конфигурация не была загружена")

	log.Printf("Подключение к %s:%d ...", cfg.IP, cfg.Port)
	c, err := fanuc.New(cfg)
	require.NoError(t, err, "Не удалось создать FOCAS клиент")
	require.NotNil(t, c, "Клиент не должен быть nil")
	log.Println("Успешно подключено!")

	return c
}

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
		log.Fatalf("Ошибка получения информации о программе: %v", err)
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

	logAsJSON(t, "MachineState (with Alarms)", state)
}

func TestReadAlarms(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	alarms, err := c.GetAlarms()
	require.NoError(t, err, "Не удалось прочитать ошибки")

	logAsJSON(t, "Alarms (standalone)", alarms)
}

func TestReadFeedData(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	feedInfo, err := c.GetFeedData()
	require.NoError(t, err, "Не удалось прочитать информацию о подаче")

	logAsJSON(t, "FeedData", feedInfo)
}

func TestReadContourFeedRate(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	contourFeedRate, err := c.GetContourFeedRate()
	require.NoError(t, err, "Не удалось прочитать информацию о контурной подаче")

	logAsJSON(t, "ContourFeedRate", contourFeedRate)
}

func TestReadFeedOverride(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	feedOverride, err := c.GetFeedOverride()
	require.NoError(t, err, "Не удалось прочитать информацию о коррекции подачи")

	logAsJSON(t, "FeedOverride", feedOverride)
}

func TestReadJogOverride(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	jogOverride, err := c.GetJogOverride()
	require.NoError(t, err, "Не удалось прочитать информацию о коррекции JOG")

	logAsJSON(t, "JogOverride", jogOverride)
}

func TestReadParameters(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	paramInfo, err := c.GetParameterInfo()
	require.NoError(t, err, "Не удалось прочитать информацию о параметрах")

	logAsJSON(t, "ParameterInfo (standalone)", paramInfo)
}

func TestGetControlProgram(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	gcode, err := c.GetControlProgram()
	if err != nil {
		log.Fatalf("Ошибка при получении содержимого программы: %v", err)
	}

	filePath := "g_code.log"
	err = os.WriteFile(filePath, []byte(gcode), 0644)
	require.NoError(t, err, "Не удалось записать G-код в файл %s", filePath)

	log.Printf("G-код программы успешно сохранен в %s", filePath)
}

func TestGetCurrentData(t *testing.T) {
	c := setupTest(t)
	defer c.Close()

	data, err := c.GetCurrentData()
	require.NoError(t, err, "Не удалось получить агрегированные данные")

	logAsJSON(t, "Aggregated Current Data", data)
}
