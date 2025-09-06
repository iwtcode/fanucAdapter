package fanuc_machine

import (
	"github.com/iwtcode/fanucService/internal/interfaces"
	"gorm.io/gorm"
)

type FanucMachineRepositoryImpl struct {
	db *gorm.DB
}

func NewFanucMachineRepository(db *gorm.DB) interfaces.FanucMachineRepository {
	return &FanucMachineRepositoryImpl{db: db}
}
