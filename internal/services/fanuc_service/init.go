package fanuc_service

import (
	"time"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/middleware/logging"
)

type fanucService struct {
	connMgr *ConnectionManager
	pollMgr *PollingManager
}

func NewFanucService(repo interfaces.FanucMachineRepository, producer interfaces.KafkaService, logger *logging.Logger) interfaces.FanucService {
	pollingManager := NewPollingManager(repo, producer, logger)
	connectionManager := NewConnectionManager(pollingManager, repo, logger)

	return &fanucService{
		connMgr: connectionManager,
		pollMgr: pollingManager,
	}
}

// --- Реализация методов интерфейса FanucService ---

func (s *fanucService) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	return s.connMgr.CreateConnection(req)
}

func (s *fanucService) RestoreConnection(machine entities.FanucMachine) (*models.ConnectionInfo, error) {
	return s.connMgr.RestoreConnection(machine)
}

func (s *fanucService) GetConnection(sessionID string) (*models.ConnectionInfo, bool) {
	return s.connMgr.GetConnection(sessionID)
}

func (s *fanucService) GetAllConnections() []*models.ConnectionInfo {
	return s.connMgr.GetAllConnections()
}

func (s *fanucService) DeleteConnection(sessionID string) error {
	return s.connMgr.DeleteConnection(sessionID)
}

func (s *fanucService) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	return s.connMgr.CheckConnection(sessionID)
}

func (s *fanucService) StartPolling(conn *models.ConnectionInfo, interval time.Duration) error {
	return s.pollMgr.StartPolling(conn, interval)
}

func (s *fanucService) StopPolling(sessionID string) error {
	return s.pollMgr.StopPolling(sessionID)
}

func (s *fanucService) IsPollingActive(sessionID string) bool {
	return s.pollMgr.IsPollingActive(sessionID)
}
