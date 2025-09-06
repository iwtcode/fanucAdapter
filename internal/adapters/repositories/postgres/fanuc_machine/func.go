package fanuc_machine

import (
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"gorm.io/gorm"
)

func (r *FanucMachineRepositoryImpl) Create(machine *entities.FanucMachine) error {
	return r.db.Create(machine).Error
}

func (r *FanucMachineRepositoryImpl) GetByEndpoint(endpointURL string) (*entities.FanucMachine, error) {
	var machine entities.FanucMachine
	err := r.db.Where("endpoint_url = ?", endpointURL).First(&machine).Error
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

// UpdatePollingState обновляет статус и интервал опроса
func (r *FanucMachineRepositoryImpl) UpdatePollingState(sessionID, status string, interval int) error {
	updates := map[string]interface{}{
		"status":   status,
		"interval": interval,
	}
	result := r.db.Model(&entities.FanucMachine{}).Where("session_id = ?", sessionID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FanucMachineRepositoryImpl) Delete(sessionID string) error {
	result := r.db.Where("session_id = ?", sessionID).Delete(&entities.FanucMachine{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *FanucMachineRepositoryImpl) GetBySessionID(sessionID string) (*entities.FanucMachine, error) {
	var machine entities.FanucMachine
	err := r.db.Where("session_id = ?", sessionID).First(&machine).Error
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

// GetAll возвращает все сохраненные станки
func (r *FanucMachineRepositoryImpl) GetAll() ([]entities.FanucMachine, error) {
	var machines []entities.FanucMachine
	if err := r.db.Find(&machines).Error; err != nil {
		return nil, err
	}
	return machines, nil
}
