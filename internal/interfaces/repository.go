package interfaces

import (
	"github.com/iwtcode/fanucService/internal/domain/entities"
)

// FanucMachineRepository определяет контракт для работы с сохраненными станками в БД
type FanucMachineRepository interface {
	Create(machine *entities.FanucMachine) error
	GetByEndpoint(endpointURL string) (*entities.FanucMachine, error)
	UpdatePollingState(sessionID, status string, interval int) error
	Delete(sessionID string) error
	GetBySessionID(sessionID string) (*entities.FanucMachine, error)
	GetAll() ([]entities.FanucMachine, error)
}
