package fanuc

import (
	"os"
	"strconv"
)

// Config хранит модель конфигурации приложения
type Config struct {
	IP          string
	Port        uint16
	TimeoutMs   int32
	ModelSeries string
	LogLevel    string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	ip := os.Getenv("FANUC_IP")
	if ip == "" {
		ip = "10.0.0.1"
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

	modelSeries := os.Getenv("FANUC_MODEL_SERIES")

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		IP:          ip,
		Port:        uint16(port),
		TimeoutMs:   int32(timeout),
		ModelSeries: modelSeries,
		LogLevel:    logLevel,
	}
}
