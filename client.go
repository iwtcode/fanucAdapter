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
	handle  uint16
	sysInfo *models.SystemInfo
	config  *Config // Используем Config из этого же пакета
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

	handle, err := focas.Connect(cfg.IP, cfg.Port, cfg.TimeoutMs)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Сразу после подключения получаем системную информацию,
	// так как она нужна для других вызовов (например, для осей).
	sysInfo, err := focas.ReadSystemInfo(handle)
	if err != nil {
		focas.Disconnect(handle) // Закрываем соединение, если не удалось получить базовую информацию
		return nil, fmt.Errorf("failed to read system info after connecting: %w", err)
	}

	return &Client{
		handle:  handle,
		sysInfo: sysInfo,
		config:  cfg,
	}, nil
}

// Close закрывает соединение со станком.
func (c *Client) Close() {
	focas.Disconnect(c.handle)
}

// GetSystemInfo возвращает системную информацию о станке.
func (c *Client) GetSystemInfo() *models.SystemInfo {
	return c.sysInfo
}

// GetMachineState возвращает текущее состояние станка.
func (c *Client) GetMachineState() (*models.UnifiedMachineData, error) {
	return focas.ReadMachineState(c.handle)
}

// GetAxisData возвращает информацию обо всех управляемых осях.
func (c *Client) GetAxisData() ([]models.AxisInfo, error) {
	if c.sysInfo == nil {
		return nil, fmt.Errorf("system info is not available")
	}
	return focas.ReadAxisData(c.handle, c.sysInfo.ControlledAxes, c.sysInfo.MaxAxis)
}

// GetSpindleData возвращает информацию обо всех шпинделях.
func (c *Client) GetSpindleData() ([]models.SpindleInfo, error) {
	return focas.ReadSpindleData(c.handle)
}

// GetProgramInfo возвращает информацию о текущей выполняемой программе.
func (c *Client) GetProgramInfo() (*models.ProgramInfo, error) {
	return focas.ReadProgram(c.handle)
}

// GetControlProgram возвращает полный G-код текущей выполняемой программы.
func (c *Client) GetControlProgram() (string, error) {
	return focas.GetControlProgram(c.handle)
}
