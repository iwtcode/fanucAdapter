package interpreter

// Model0iInterpreter может переопределять или расширять поведение интерпретатора по умолчанию.
// В данный момент он просто встраивает неизвестную модель для наследования ее методов.
type Model15iInterpreter struct {
	ModelUnknownInterpreter
}
