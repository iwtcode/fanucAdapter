package focas

import (
	"strings"

	"github.com/iwtcode/fanucService/focas/interpreter"
	"github.com/iwtcode/fanucService/focas/model"
	"github.com/iwtcode/fanucService/focas/program"
)

// GetModelImplementations выбирает подходящие интерпретатор и ридер программ
// на основе строки серии ЧПУ.
func GetModelImplementations(series string) (model.Interpreter, model.ProgramReader) {
	s := strings.ToUpper(series)

	// Реализации по умолчанию
	var interp model.Interpreter = &interpreter.ModelUnknownInterpreter{}
	var progReader model.ProgramReader = &program.ModelUnknownProgramReader{}

	// Переопределения для конкретных моделей
	if strings.HasPrefix(s, "0I") {
		interp = &interpreter.Model0iInterpreter{}
		progReader = &program.Model0iProgramReader{}
	} else if strings.HasPrefix(s, "15I") {
		interp = &interpreter.Model15iInterpreter{}
		progReader = &program.Model15iProgramReader{}
	} else if strings.HasPrefix(s, "15") {
		interp = &interpreter.Model15Interpreter{}
		progReader = &program.Model15ProgramReader{}
	} else if strings.HasPrefix(s, "16I") {
		interp = &interpreter.Model16iInterpreter{}
		progReader = &program.Model16iProgramReader{}
	} else if strings.HasPrefix(s, "16") {
		interp = &interpreter.Model16Interpreter{}
		progReader = &program.Model16ProgramReader{}
	} else if strings.HasPrefix(s, "18I") {
		interp = &interpreter.Model18iInterpreter{}
		progReader = &program.Model18iProgramReader{}
	} else if strings.HasPrefix(s, "21") {
		interp = &interpreter.Model21Interpreter{}
		progReader = &program.Model21ProgramReader{}
	} else if strings.HasPrefix(s, "30") {
		interp = &interpreter.Model30Interpreter{}
		progReader = &program.Model30ProgramReader{}
	} else if strings.HasPrefix(s, "31") {
		interp = &interpreter.Model31Interpreter{}
		progReader = &program.Model31ProgramReader{}
	} else if strings.HasPrefix(s, "32") {
		interp = &interpreter.Model32Interpreter{}
		progReader = &program.Model32ProgramReader{}
	}

	return interp, progReader
}
