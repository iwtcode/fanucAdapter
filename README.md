<div align="center">

# Fanuc Focas Adapter

![alt text](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![alt text](https://img.shields.io/badge/Fanuc-Focas-yellow)
![alt text](https://img.shields.io/badge/CGO-Bindings-blue)
![alt text](https://img.shields.io/badge/License-MIT-green)

*Go-обертка для библиотеки Fanuc FOCAS (libfwlib32), обеспечивающая удобное взаимодействие со станками ЧПУ*

</div>

### ✨ Ключевые возможности

- 🚀 **Получение состояний**: Чтение состояния станка, осей, шпинделей, ошибок, программ и параметров.
- 🔌 **Устойчивое соединение**: Встроенная логика автоматического переподключения (`Reconnect`) при разрыве связи.
- 🧵 **Потокобезопасность**: Глобальная синхронизация вызовов C-библиотеки для безопасной работы в конкурентной среде.
- 🏭 **Мультимодельная поддержка**: Фабричный метод инициализации для различных серий Fanuc (0i, 16i, 30i, 31i и др.).
- 📦 **Агрегация данных**: Метод `GetCurrentData` для получения полного состояния станка одним вызовом.
- 🛠️ **CGO Bindings**: Низкоуровневая интеграция с нативной библиотекой `libfwlib32`.

## 🏗️ Архитектура

Библиотека выступает прослойкой между Go-приложением и нативной C-библиотекой Fanuc.

```
┌─────────────────┐      ┌─────────────────────────┐      ┌──────────────────┐
│  Go Application │      │      fanucAdapter       │      │   Focas Library  │
│    (Client)     ├─────▸│       (Wrapper)         ├─────▸│   (libfwlib32)   │
└─────────────────┘      │   - Auto Reconnect      │      │     (.dll / .so) │
                         │   - Thread Safety (Lock)│      └─────────┬────────┘
                         │   - Data Mapping        │                │ TCP/IP
                         └─────────────────────────┘                │
                                                                    ▾
                                                          ┌──────────────────┐
                                                          │    Fanuc CNC     │
                                                          │   Machine Tool   │
                                                          └──────────────────┘
```

## 📦 Установка

```bash
go get github.com/iwtcode/fanucAdapter
```

## 🚀 Использование

### Инициализация клиента

```go
package main

import (
	"fmt"
	"log"

	fanuc "github.com/iwtcode/fanucAdapter"
	"github.com/iwtcode/fanucAdapter/models"
)

func main() {
	// Конфигурация подключения
	cfg := &fanuc.Config{
		IP:          "10.0.0.1",
		Port:        8193,
		TimeoutMs:   5000,
		ModelSeries: "0i",
		LogLevel:    "info",
	}

	// Создание клиента
	client, err := fanuc.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

    // Получение агрегированных данных
    data, err := client.GetCurrentData()
    if err != nil {
        log.Println("Error:", err)
        return
    }

    fmt.Printf("Станок: %s\n", data.MachineID)
    fmt.Printf("Режим: %s, Статус: %s\n", data.TmMode, data.MachineState)
    // ...
}
```

## 🔧 Конфигурация

Библиотеку можно настроить через структуру `fanuc.Config` или переменные окружения (при использовании `fanuc.Load()`).

| Переменная окружения | Параметр в Config | Описание | Значение по умолчанию |
|----------------------|-------------------|----------|-----------------------|
| `FANUC_IP` | `IP` | IP адрес станка | `10.0.0.1` |
| `FANUC_PORT` | `Port` | Focas порт | `8193` |
| `FANUC_TIMEOUT` | `TimeoutMs` | Таймаут соединения (мс) | `5000` |
| `FANUC_MODEL_SERIES` | `ModelSeries` | Серия станка | `Unknown` |
| `LOG_LEVEL` | `LogLevel` | Уровень логирования | `info` |

## 📁 Структура проекта

```
fanucAdapter/
├── client.go           # Публичный API клиента
├── config.go           # Загрузка конфигурации
├── models/             # Структуры данных (DTO)
├── focas/              # Внутренняя реализация
└── tests/              # Интеграционные тесты
```

## 🧪 Тестирование

```bash
go test -v -count=1 ./tests
```

## 📝 Лицензия

Проект распространяется под [лицензией MIT](LICENSE).

---
Copyright (c) 2025 iwtcode