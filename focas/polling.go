package focas

import (
	"context"
	"log"
	"time"

	"github.com/iwtcode/fanucService/models"
)

// PollingResult содержит данные или ошибку от одной попытки опроса.
type PollingResult struct {
	Data *models.AggregatedData
	Err  error
}

// StartPolling запускает фоновый процесс, который периодически вызывает AggregateAllData.
// Он обрабатывает одновременные запросы, если сбор данных занимает больше времени, чем интервал.
// Опрос прекращается при отмене предоставленного контекста.
func (a *FocasAdapter) StartPolling(ctx context.Context, interval time.Duration) <-chan PollingResult {
	resultsChan := make(chan PollingResult)

	go func() {
		defer close(resultsChan)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Опрос остановлен из-за отмены контекста.")
				return
			case <-ticker.C:
				go func() {
					data, err := a.AggregateAllData()
					result := PollingResult{Data: data, Err: err}
					select {
					case resultsChan <- result:
					case <-ctx.Done():
					}
				}()
			}
		}
	}()

	return resultsChan
}
