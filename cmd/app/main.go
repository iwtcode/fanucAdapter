// @title Fanuc Service API
// @version 1.0.0
// @description API для работы со станками Fanuc по протоколу FOCAS2 и отправки данных в Kafka.
// @host localhost:8080
// @BasePath /api/v1
package main

import "github.com/iwtcode/fanucService/internal/app"

func main() {
	// Создаем и запускаем новый экземпляр приложения fx
	app.New().Run()
}
