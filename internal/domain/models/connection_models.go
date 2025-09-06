package models

import "time"

// ConnectionRequest определяет структуру для нового запроса на подключение.
type ConnectionRequest struct {
	EndpointURL string `json:"endpoint_url" binding:"required"` // "192.168.1.10:8193"
}

// SessionRequest определяет структуру для запросов, использующих SessionID.
type SessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// PollingRequest определяет структуру для запроса на запуск опроса.
type PollingRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Interval  int    `json:"interval" binding:"required,gt=0"` // в миллисекундах
}

// ConnectionInfo представляет активное подключение в пуле.
type ConnectionInfo struct {
	SessionID string    `json:"session_id"`
	Endpoint  string    `json:"endpoint"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
	UseCount  int64     `json:"use_count"`
	IsHealthy bool      `json:"is_healthy"`
}
