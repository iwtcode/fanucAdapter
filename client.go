package fanuc

import (
	"fmt"
	"sync"

	"github.com/iwtcode/fanucService/focas"
	"github.com/iwtcode/fanucService/models"
)

var (
	startupOnce sync.Once
	startupErr  error
)

// Client является основной точкой входа для взаимодействия с библиотекой.
type Client struct {
	adapter *focas.FocasAdapter
	config  *Config
}

// New создает и возвращает новый экземпляр клиента.
// Эта функция инициализирует FOCAS (только один раз) и устанавливает соединение.
func New(cfg *Config) (*Client, error) {
	// Инициализация FOCAS должна происходить только один раз за все время работы приложения.
	startupOnce.Do(func() {
		// Используем режим 3 для логирования в файл
		startupErr = focas.Startup(3, cfg.LogPath)
	})

	if startupErr != nil {
		return nil, fmt.Errorf("FOCAS startup failed: %w", startupErr)
	}

	// Передаем указанную серию модели в адаптер
	adapter, err := focas.NewFocasAdapter(cfg.IP, cfg.Port, cfg.TimeoutMs, cfg.ModelSeries)
	if err != nil {
		return nil, fmt.Errorf("failed to create focas adapter: %w", err)
	}

	return &Client{
		adapter: adapter,
		config:  cfg,
	}, nil
}

// Close закрывает соединение со станком.
func (c *Client) Close() {
	if c.adapter != nil {
		c.adapter.Close()
	}
}

// GetSystemInfo возвращает системную информацию о станке.
func (c *Client) GetSystemInfo() *models.SystemInfo {
	return c.adapter.GetSystemInfo()
}

// GetMachineState возвращает текущее состояние станка.
func (c *Client) GetMachineState() (*models.UnifiedMachineData, error) {
	return c.adapter.ReadMachineState()
}

// GetAxisData возвращает информацию обо всех управляемых осях.
func (c *Client) GetAxisData() ([]models.AxisInfo, error) {
	return c.adapter.ReadAxisData()
}

// GetSpindleData возвращает информацию обо всех шпинделях.
func (c *Client) GetSpindleData() ([]models.SpindleInfo, error) {
	return c.adapter.ReadSpindleData()
}

// GetProgramInfo возвращает информацию о текущей выполняемой программе.
func (c *Client) GetProgramInfo() (*models.ProgramInfo, error) {
	return c.adapter.ReadProgram()
}

// GetControlProgram возвращает полный G-код текущей выполняемой программы.
func (c *Client) GetControlProgram() (string, error) {
	return c.adapter.GetControlProgram()
}

// GetAlarms возвращает список активных ошибок на станке.
func (c *Client) GetAlarms() ([]models.AlarmDetail, error) {
	return c.adapter.ReadAlarms()
}

// GetFeedData возвращает информацию о скорости подачи и коррекции.
func (c *Client) GetFeedData() (*models.FeedInfo, error) {
	return c.adapter.ReadFeedData()
}

// GetContourFeedRate возвращает фактическую скорость подачи по контуру.
func (c *Client) GetContourFeedRate() (int32, error) {
	return c.adapter.ReadContourFeedRate()
}

// GetFeedOverride возвращает процент коррекции подачи.
func (c *Client) GetFeedOverride() (int32, error) {
	return c.adapter.ReadFeedOverride()
}

// GetJogOverride возвращает процент коррекции скорости в режиме JOG.
func (c *Client) GetJogOverride() (int32, error) {
	return c.adapter.ReadJogOverride()
}

// GetParameterInfo возвращает информацию о параметрах (счетчики, время работы).
func (c *Client) GetParameterInfo() (*models.ParameterInfo, error) {
	return c.adapter.ReadParameterInfo()
}

// GetCurrentData возвращает полную сводку данных о станке, собранную асинхронно.
func (c *Client) GetCurrentData() (*models.AggregatedData, error) {
	return c.adapter.AggregateAllData()
}
