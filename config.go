package fanuc

import (
	"os"
	"strconv"
)

// Config хранит модель конфигурации приложения
type Config struct {
	IP        string
	Port      uint16
	TimeoutMs int32
	LogPath   string
}

// Load загружает конфигурацию из переменных окружения или устанавливает значения по умолчанию
func Load() *Config {
	logPath := os.Getenv("FANUC_LOG_PATH")
	if logPath == "" {
		logPath = "./fanuc.log"
	}

	ip := os.Getenv("FANUC_IP")
	if ip == "" {
		ip = "192.168.0.3"
	}

	portStr := os.Getenv("FANUC_PORT")
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil || port == 0 {
		port = 8193
	}

	timeoutStr := os.Getenv("FANUC_TIMEOUT")
	timeout, err := strconv.ParseInt(timeoutStr, 10, 32)
	if err != nil || timeout == 0 {
		timeout = 5000
	}

	return &Config{
		IP:        ip,
		Port:      uint16(port),
		TimeoutMs: int32(timeout),
		LogPath:   logPath,
	}
}
