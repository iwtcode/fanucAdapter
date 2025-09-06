package usecases

import "github.com/iwtcode/fanucService/internal/interfaces"

// UseCases - агрегатор всех use case интерфейсов
type UseCases struct {
	interfaces.Usecases
}

// NewUsecases - конструктор для UseCases
func NewUsecases(
	fanucSvc interfaces.FanucService,
) interfaces.Usecases {
	return NewUsecase(fanucSvc)
}
