package fanuc_service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/middleware/logging"
	"github.com/iwtcode/fanucService/internal/services/fanuc_service/focas"
)

type activePoll struct {
	ticker *time.Ticker
	done   chan bool
}

type PollingManager struct {
	dbRepo      interfaces.FanucMachineRepository
	producer    interfaces.KafkaService
	logger      *logging.Logger
	activePolls map[string]*activePoll
	pollsMutex  sync.Mutex
}

func NewPollingManager(dbRepo interfaces.FanucMachineRepository, producer interfaces.KafkaService, logger *logging.Logger) *PollingManager {
	return &PollingManager{
		dbRepo:      dbRepo,
		producer:    producer,
		logger:      logger.WithPrefix("POLLER"),
		activePolls: make(map[string]*activePoll),
	}
}

func (pm *PollingManager) IsPollingActive(sessionID string) bool {
	pm.pollsMutex.Lock()
	defer pm.pollsMutex.Unlock()
	_, exists := pm.activePolls[sessionID]
	return exists
}

func (pm *PollingManager) StartPolling(conn *models.ConnectionInfo, interval time.Duration) error {
	pm.pollsMutex.Lock()
	defer pm.pollsMutex.Unlock()

	sessionID := conn.SessionID
	if _, exists := pm.activePolls[sessionID]; exists {
		return fmt.Errorf("опрос для сессии '%s' уже запущен", sessionID)
	}

	if err := pm.dbRepo.UpdatePollingState(sessionID, entities.StatusPolled, int(interval.Milliseconds())); err != nil {
		return fmt.Errorf("не удалось обновить статус станка в БД: %w", err)
	}

	pm.startPollingForMachineUnsafe(conn.SessionID, conn.Endpoint, interval)
	return nil
}

func (pm *PollingManager) StopPolling(sessionID string) error {
	pm.pollsMutex.Lock()
	defer pm.pollsMutex.Unlock()

	if err := pm.dbRepo.UpdatePollingState(sessionID, entities.StatusConnected, 0); err != nil {
		pm.logger.Error("Failed to update status in DB when stopping polling", "sessionID", sessionID, "error", err)
	}

	pm.stopPollingUnsafe(sessionID)
	return nil
}

func (pm *PollingManager) StopPollingForMachine(sessionID string) {
	pm.pollsMutex.Lock()
	defer pm.pollsMutex.Unlock()
	pm.stopPollingUnsafe(sessionID)
}

func (pm *PollingManager) stopPollingUnsafe(sessionID string) {
	poll, exists := pm.activePolls[sessionID]
	if !exists {
		return
	}
	poll.ticker.Stop()
	poll.done <- true
	close(poll.done)
	delete(pm.activePolls, sessionID)
	pm.logger.Info("Polling stopped", "sessionID", sessionID)
}

func (pm *PollingManager) startPollingForMachineUnsafe(sessionID, endpoint string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	done := make(chan bool)

	pm.activePolls[sessionID] = &activePoll{
		ticker: ticker,
		done:   done,
	}

	go func() {
		pm.logger.Info("Starting polling goroutine", "sessionID", sessionID, "endpoint", endpoint, "interval", interval)

		host, portStr, _ := net.SplitHostPort(endpoint)
		port, _ := strconv.Atoi(portStr)

		defer func() {
			pm.logger.Info("Polling goroutine stopped", "sessionID", sessionID)
		}()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// Шаг 1: Устанавливаем соединение в начале каждой итерации
				handle, err := focas.Connect(host, uint16(port), 5000)
				if err != nil {
					pm.logger.Error("Failed to connect for polling tick", "sessionID", sessionID, "error", err)
					continue // Пропускаем эту итерацию, попробуем на следующей
				}
				pm.logger.Debug("Successfully connected for polling tick", "sessionID", sessionID, "handle", handle)

				// Шаг 2: Получаем данные
				machineData, err := focas.GetAllData(handle, sessionID, endpoint)

				// Шаг 3: Всегда разрываем соединение, независимо от результата
				focas.Disconnect(handle)
				pm.logger.Debug("Disconnected from polling connection", "sessionID", sessionID)

				// Шаг 4: Обрабатываем результат после разрыва соединения
				if err != nil {
					pm.logger.Error("Error getting machine data", "sessionID", sessionID, "error", err)
					continue // Пропускаем эту итерацию
				}

				jsonData, err := json.Marshal(machineData)
				if err != nil {
					pm.logger.Error("Failed to serialize data for Kafka", "sessionID", sessionID, "error", err)
					continue
				}

				err = pm.producer.Produce(context.Background(), []byte(sessionID), jsonData)
				if err != nil {
					pm.logger.Error("Failed to send data to Kafka", "sessionID", sessionID, "error", err)
				}
			}
		}
	}()
}
