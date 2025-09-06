package fanuc_service

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/middleware/logging"
	"github.com/iwtcode/fanucService/internal/services/fanuc_service/focas"
	"gorm.io/gorm"
)

// PollingStarter определяет методы, которые ConnectionManager может вызывать у PollingManager.
type PollingStarter interface {
	StopPollingForMachine(sessionID string)
}

type ConnectionManager struct {
	mu         sync.RWMutex
	pool       map[string]*models.ConnectionInfo
	pollingMgr PollingStarter // Используем интерфейс
	dbRepo     interfaces.FanucMachineRepository
	logger     *logging.Logger
}

func NewConnectionManager(pollingMgr PollingStarter, dbRepo interfaces.FanucMachineRepository, logger *logging.Logger) *ConnectionManager {
	return &ConnectionManager{
		pool:       make(map[string]*models.ConnectionInfo),
		pollingMgr: pollingMgr,
		dbRepo:     dbRepo,
		logger:     logger.WithPrefix("CONNECTOR"),
	}
}

func (cm *ConnectionManager) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	_, _, err := net.SplitHostPort(req.EndpointURL)
	if err != nil {
		return nil, fmt.Errorf("неверный формат endpoint_url. Ожидается 'IP:PORT', получено '%s'", req.EndpointURL)
	}

	existingMachineDB, err := cm.dbRepo.GetByEndpoint(req.EndpointURL)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ошибка при проверке станка в БД: %w", err)
	}
	if existingMachineDB != nil {
		cm.mu.RLock()
		_, exists := cm.pool[existingMachineDB.SessionID]
		cm.mu.RUnlock()
		if exists {
			return nil, fmt.Errorf("подключение для '%s' уже активно с SessionID: %s", req.EndpointURL, existingMachineDB.SessionID)
		}
		cm.logger.Warn("Connection for endpoint exists in DB but not in pool. Deleting old DB record and creating a new session.", "endpoint", req.EndpointURL)
		_ = cm.dbRepo.Delete(existingMachineDB.SessionID)
	}

	if err := cm.checkMachineConnection(req.EndpointURL); err != nil {
		return nil, fmt.Errorf("первичная проверка подключения провалена: %w", err)
	}

	sessionID := uuid.New().String()
	machineToSave := &entities.FanucMachine{
		SessionID:   sessionID,
		EndpointURL: req.EndpointURL,
		Status:      entities.StatusConnected,
	}
	if err := cm.dbRepo.Create(machineToSave); err != nil {
		return nil, fmt.Errorf("не удалось сохранить новое подключение %s в БД: %w", sessionID, err)
	}

	connInfo := &models.ConnectionInfo{
		SessionID: sessionID,
		Endpoint:  req.EndpointURL,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
		IsHealthy: true,
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.pool[sessionID] = connInfo

	cm.logger.Info("Connection created successfully", "sessionID", sessionID, "endpoint", req.EndpointURL)
	return connInfo, nil
}

func (cm *ConnectionManager) RestoreConnection(machine entities.FanucMachine) (*models.ConnectionInfo, error) {
	connInfo := &models.ConnectionInfo{
		SessionID: machine.SessionID,
		Endpoint:  machine.EndpointURL,
		CreatedAt: machine.CreatedAt,
		LastUsed:  time.Now(),
		IsHealthy: false, // По умолчанию нездоровое, пока не проверим
	}

	err := cm.checkMachineConnection(machine.EndpointURL)
	connInfo.IsHealthy = (err == nil)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.pool[machine.SessionID] = connInfo

	return connInfo, nil
}

func (cm *ConnectionManager) GetConnection(sessionID string) (*models.ConnectionInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, found := cm.pool[sessionID]
	return conn, found
}

func (cm *ConnectionManager) GetAllConnections() []*models.ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conns := make([]*models.ConnectionInfo, 0, len(cm.pool))
	for _, conn := range cm.pool {
		conns = append(conns, conn)
	}
	return conns
}

func (cm *ConnectionManager) DeleteConnection(sessionID string) error {
	// Сначала останавливаем опрос, если он был
	cm.pollingMgr.StopPollingForMachine(sessionID)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.pool[sessionID]; !exists {
		err := cm.dbRepo.Delete(sessionID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("ошибка удаления сессии '%s' из БД: %w", sessionID, err)
		}
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("сессия '%s' не найдена ни в активном пуле, ни в БД", sessionID)
		}
		cm.logger.Info("Session (not in pool) successfully deleted from DB.", "sessionID", sessionID)
		return nil
	}

	delete(cm.pool, sessionID)

	if err := cm.dbRepo.Delete(sessionID); err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("ошибка удаления сессии '%s' из БД: %w", sessionID, err)
	}

	cm.logger.Info("Session deleted successfully.", "sessionID", sessionID)
	return nil
}

func (cm *ConnectionManager) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.pool[sessionID]
	if !exists {
		return nil, fmt.Errorf("сессия '%s' не найдена", sessionID)
	}

	previousHealth := conn.IsHealthy
	err := cm.checkMachineConnection(conn.Endpoint)
	conn.IsHealthy = (err == nil)
	conn.LastUsed = time.Now()
	conn.UseCount++

	if previousHealth != conn.IsHealthy {
		cm.logger.Info("Session health status changed", "sessionID", sessionID, "from", previousHealth, "to", conn.IsHealthy)
	}

	return conn, err
}

// checkMachineConnection - это внутренняя реализация проверки, она не часть интерфейса PollingStarter
func (cm *ConnectionManager) checkMachineConnection(endpoint string) error {
	host, portStr, _ := net.SplitHostPort(endpoint)
	port, _ := strconv.Atoi(portStr)

	handle, err := focas.Connect(host, uint16(port), 2000) // Короткий таймаут для проверки
	if err != nil {
		return err
	}
	focas.Disconnect(handle)
	return nil
}
