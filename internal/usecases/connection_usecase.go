package usecases

import (
	"fmt"
	"time"

	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/domain/models"
	"github.com/iwtcode/fanucService/internal/interfaces"
)

type Usecase struct {
	fanucSvc interfaces.FanucService
}

func NewUsecase(fanucSvc interfaces.FanucService) interfaces.Usecases {
	return &Usecase{
		fanucSvc: fanucSvc,
	}
}

func (u *Usecase) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	return u.fanucSvc.CreateConnection(req)
}

func (u *Usecase) RestoreConnection(machine entities.FanucMachine) (*models.ConnectionInfo, error) {
	return u.fanucSvc.RestoreConnection(machine)
}

func (u *Usecase) GetAllConnections() []*models.ConnectionInfo {
	return u.fanucSvc.GetAllConnections()
}

func (u *Usecase) DeleteConnection(sessionID string) error {
	return u.fanucSvc.DeleteConnection(sessionID)
}

func (u *Usecase) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	return u.fanucSvc.CheckConnection(sessionID)
}

func (u *Usecase) StartPolling(sessionID string, interval time.Duration) error {
	conn, found := u.fanucSvc.GetConnection(sessionID)
	if !found {
		return fmt.Errorf("не удалось запустить опрос: сессия '%s' не найдена в активном пуле", sessionID)
	}
	return u.fanucSvc.StartPolling(conn, interval)
}

func (u *Usecase) StopPolling(sessionID string) error {
	return u.fanucSvc.StopPolling(sessionID)
}
