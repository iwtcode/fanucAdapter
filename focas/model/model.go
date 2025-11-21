package model

import (
	"unsafe"

	"github.com/iwtcode/fanucAdapter/models"
)

// FanucModel представляет конкретную серию моделей ЧПУ.
type FanucModel int

const (
	Model_Unknown FanucModel = iota
	Model_0i
	Model_15
	Model_15i
	Model_16
	Model_16i
	Model_18i
	Model_21
	Model_30
	Model_31
	Model_32
)

// Константы для строковых представлений серий моделей ЧПУ
const (
	Series0i  = "0I"
	Series15  = "15"
	Series15i = "15I"
	Series16  = "16"
	Series16i = "16I"
	Series18i = "18I"
	Series21  = "21"
	Series30  = "30"
	Series31  = "31"
	Series32  = "32"
)

// Interpreter определяет интерфейс для интерпретации состояния станка в зависимости от модели.
type Interpreter interface {
	InterpretMachineState(stat unsafe.Pointer) *models.UnifiedMachineData
}

// ProgramReader определяет интерфейс для логики чтения управляющей программы в зависимости от модели.
// Ему необходим доступ к адаптеру для выполнения вызовов FOCAS.
type ProgramReader interface {
	GetControlProgram(adapter FocasCaller) (string, error)
}

// FocasCaller - это интерфейс, который абстрагирует FocasAdapter,
// предоставляя только те методы, которые необходимы для реализаций ProgramReader.
// Это предотвращает циклические зависимости между пакетом program и пакетом focas.
type FocasCaller interface {
	ReadProgram() (*models.ProgramInfo, error)
	CallWithReconnect(f func(handle uint16) (int16, error)) error
}
