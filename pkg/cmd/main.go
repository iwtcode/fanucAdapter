package main

import (
	"fmt"
	"log"

	"github.com/iwtcode/fanucService/pkg/config"
	"github.com/iwtcode/fanucService/pkg/focas"
)

func main() {
	// 1) Загрузка конфигурации
	cfg := config.Load()

	// 2) Инициализация FOCAS2 процесса + лог
	if err := focas.Startup(3, cfg.LogPath); err != nil {
		log.Fatalf("FOCAS startup error: %v", err)
	}

	// 3) Подключение к ЧПУ
	fmt.Printf("Подключение к %s:%d ...\n", cfg.IP, cfg.Port)
	h, err := focas.Connect(cfg.IP, cfg.Port, cfg.TimeoutMs)
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer focas.Disconnect(h)
	fmt.Println("Успешно подключено!")

	// 4) Чтение системной информации
	fmt.Println("\n--- Системная информация ---")
	systemInfo, err := focas.ReadSystemInfo(h)
	if err != nil {
		log.Fatalf("Не удалось прочитать системную информацию: %v", err)
	}
	fmt.Printf("Производитель: %s\n", systemInfo.Manufacturer)
	fmt.Printf("Модель:        %s\n", systemInfo.Model)
	fmt.Printf("Серия:         %s\n", systemInfo.Series)
	fmt.Printf("Версия:        %s\n", systemInfo.Version)
	fmt.Printf("Кол-во осей:   %d\n", systemInfo.ControlledAxes)

	fmt.Println("\n--- Информация об осях ---")
	axisInfos, err := focas.ReadAxisData(h, systemInfo.ControlledAxes, systemInfo.MaxAxis)
	if err != nil {
		log.Fatalf("Не удалось прочитать информацию об осях: %v", err)
	}
	fmt.Println(`"axis_infos": [`)
	for i, axis := range axisInfos {
		fmt.Println("    {")
		fmt.Printf("        \"name\": \"%s\",\n", axis.Name)
		fmt.Printf("        \"position\": %.3f\n", axis.Position)
		if i < len(axisInfos)-1 {
			fmt.Println("    },")
		} else {
			fmt.Println("    }")
		}
	}
	fmt.Println("]")

	// 5) Чтение информации о программе
	fmt.Println("\n--- Информация о программе ---")
	progInfo, err := focas.ReadProgram(h)
	if err != nil {
		log.Printf("Не удалось прочитать информацию о программе: %v", err)
	} else {
		fmt.Printf("Имя программы: %s\n", progInfo.Name)
		fmt.Printf("Номер программы: %d\n", progInfo.Number)
	}

	// 6) Чтение полного состояния станка
	fmt.Println("\n--- Состояние станка ---")
	machineData, err := focas.ReadMachineState(h)
	if err != nil {
		log.Fatalf("Не удалось прочитать состояние станка: %v", err)
	}

	fmt.Printf("Режим (T/M):             %s\n", machineData.TmMode)
	fmt.Printf("Режим работы:            %s\n", machineData.ProgramMode)
	fmt.Printf("Состояние выполнения:    %s\n", machineData.MachineState)
	fmt.Printf("Движение осей:           %s\n", machineData.AxisMovementStatus)
	fmt.Printf("Статус M/S/T/B:          %s\n", machineData.MstbStatus)
	fmt.Printf("Статус аварийного стопа: %s\n", machineData.EmergencyStatus)
	fmt.Printf("Статус тревоги:          %s\n", machineData.AlarmStatus)
	fmt.Printf("Статус редактирования:   %s\n", machineData.EditStatus)
}
