package focas

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../ -lfwlib32

#include "fwlib32.h"
*/
import "C"

import (
	"fmt"
	"sync"
	"time"

	"github.com/iwtcode/fanucService/models"
)

// FocasAdapter инкапсулирует логику подключения и вызовов к FOCAS API.
// Он также управляет автоматическим переподключением.
type FocasAdapter struct {
	ip      string
	port    uint16
	timeout int32
	handle  uint16
	mu      sync.Mutex
	sysInfo *models.SystemInfo // ИСПРАВЛЕНО: Храним информацию о системе здесь
}

// NewFocasAdapter создает новый экземпляр FocasAdapter и устанавливает соединение.
func NewFocasAdapter(ip string, port uint16, timeoutMs int32) (*FocasAdapter, error) {
	handle, err := Connect(ip, port, timeoutMs)
	if err != nil {
		return nil, fmt.Errorf("initial connection failed: %w", err)
	}

	// Сразу после подключения получаем системную информацию
	sysInfo, err := ReadSystemInfo(handle)
	if err != nil {
		Disconnect(handle) // Закрываем соединение, если не удалось получить базовую информацию
		return nil, fmt.Errorf("failed to read system info after connecting: %w", err)
	}

	adapter := &FocasAdapter{
		ip:      ip,
		port:    port,
		timeout: timeoutMs,
		handle:  handle,
		sysInfo: sysInfo, // ИСПРАВЛЕНО: Сохраняем полученную информацию
	}

	return adapter, nil
}

// reconnect пытается восстановить соединение.
func (a *FocasAdapter) reconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.handle != 0 {
		Disconnect(a.handle)
		a.handle = 0
		time.Sleep(200 * time.Millisecond)
	}

	newHandle, err := Connect(a.ip, a.port, a.timeout)
	if err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}

	a.handle = newHandle
	fmt.Println("Successfully reconnected to FOCAS.")
	return nil
}

// D:\vs\go\fanucService\focas\adapter.go

// callWithReconnect — это обертка для выполнения вызовов с возможностью переподключения.
// При ошибках соединения функция будет бесконечно пытаться переподключиться и повторить операцию.
func (a *FocasAdapter) callWithReconnect(f func(handle uint16) (int16, error)) error {
	for {
		a.mu.Lock()
		currentHandle := a.handle
		a.mu.Unlock()

		rc, err := f(currentHandle)

		// 1. Если операция прошла успешно, выходим.
		if err == nil {
			return nil
		}

		// 2. Проверяем, является ли ошибка ошибкой соединения.
		if rc == C.EW_HANDLE || rc == C.EW_SOCKET {
			fmt.Printf("Connection error detected (rc=%d). Attempting to reconnect...\n", rc)

			// 3. Пытаемся переподключиться.
			if reconnErr := a.reconnect(); reconnErr != nil {
				// Если само переподключение не удалось, ждем и пробуем снова.
				fmt.Printf("Reconnect failed: %v. Retrying in 1 second...\n", reconnErr)
				time.Sleep(1 * time.Second)
				continue
			}

			// 4. После успешного переподключения, переходим на следующую итерацию цикла,
			// чтобы повторить исходную операцию с новым хендлом.
			// Добавляем небольшую задержку для стабилизации.
			time.Sleep(200 * time.Millisecond)
			continue

		} else {
			// 5. Если ошибка другого типа (не связана с соединением),
			// нет смысла переподключаться. Возвращаем ошибку.
			return err
		}
	}
}

// Close закрывает соединение.
func (a *FocasAdapter) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.handle != 0 {
		Disconnect(a.handle)
		a.handle = 0
	}
}

// GetSystemInfo возвращает системную информацию о станке.
// ИСПРАВЛЕНО: Просто возвращаем сохраненное значение, без повторного вызова.
func (a *FocasAdapter) GetSystemInfo() *models.SystemInfo {
	return a.sysInfo
}
