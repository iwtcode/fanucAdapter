package config

import (
	"os"
)

// Config хранит конфигурацию приложения.
type Config struct {
	IP        string
	Port      uint16
	TimeoutMs int32
	LogPath   string
}

// Load загружает конфигурацию из переменных окружения или устанавливает значения по умолчанию.
func Load() *Config {
	logPath := os.Getenv("FOCAS_LOG_PATH")
	if logPath == "" {
		logPath = "./focas2.log"
	}

	ip := os.Getenv("FANUC_IP")
	if ip == "" {
		ip = "192.168.0.6"
		// ip = "192.168.30.142"
	}

	return &Config{
		IP:        ip,
		Port:      8193,
		TimeoutMs: 5000,
		LogPath:   logPath,
	}
}
