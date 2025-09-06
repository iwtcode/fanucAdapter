package app

import (
	"context"
	"net/http"
	"time"

	"github.com/iwtcode/fanucService/internal/adapters/handlers"
	"github.com/iwtcode/fanucService/internal/adapters/repositories/postgres"
	"github.com/iwtcode/fanucService/internal/config"
	"github.com/iwtcode/fanucService/internal/domain/entities"
	"github.com/iwtcode/fanucService/internal/interfaces"
	"github.com/iwtcode/fanucService/internal/middleware/logging"
	"github.com/iwtcode/fanucService/internal/middleware/swagger"
	"github.com/iwtcode/fanucService/internal/services/fanuc_service"
	"github.com/iwtcode/fanucService/internal/services/fanuc_service/focas"
	"github.com/iwtcode/fanucService/internal/services/kafka"
	"github.com/iwtcode/fanucService/internal/usecases"

	"go.uber.org/fx"
)

// New создает новый экземпляр fx.App
func New() *fx.App {
	return fx.New(
		ConfigModule,
		LoggingModule,
		RepositoryModule,
		ProducerModule,
		ServiceModule,
		UsecaseModule,
		HttpServerModule,
		// Invoke-функции для запуска фоновых задач и хуков жизненного цикла
		fx.Invoke(InvokeInitializeFocas),
		fx.Invoke(InvokeRestoreConnections),
	)
}

// --- Модули FX ---

var ConfigModule = fx.Module("config_module",
	fx.Provide(config.LoadConfiguration),
)

func ProvideLogger(cfg *config.AppConfig) *logging.Logger {
	loggerCfg := &logging.Config{
		Enabled:    cfg.Logging.Enable,
		Level:      cfg.Logging.Level,
		LogsDir:    cfg.Logging.LogsDir,
		SavingDays: uint(cfg.Logging.SavingDays),
	}
	return logging.NewLogger(loggerCfg, "FanucServiceApp")
}

var LoggingModule = fx.Module("logging_module",
	fx.Provide(ProvideLogger),
)

var RepositoryModule = fx.Module("repository_module",
	fx.Provide(postgres.NewRepository),
)

var ProducerModule = fx.Module("producer_module",
	fx.Provide(kafka.NewKafkaProducer),
)

var ServiceModule = fx.Module("service_module",
	fx.Provide(fanuc_service.NewFanucService),
)

var UsecaseModule = fx.Module("usecases_module",
	fx.Provide(usecases.NewUsecases),
)

func NewSwaggerConfig() *swagger.Config {
	return &swagger.Config{
		Enabled: true,
		Path:    "/swagger",
	}
}

var HttpServerModule = fx.Module("http_server_module",
	fx.Provide(
		NewSwaggerConfig,
		handlers.NewHandler,
		handlers.ProvideRouter,
	),
	fx.Invoke(InvokeHttpServer),
)

// InvokeInitializeFocas инициализирует FOCAS библиотеку при старте.
func InvokeInitializeFocas(lc fx.Lifecycle, logger *logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Initializing FOCAS2 library...")
			if err := focas.InitializeFocas(); err != nil {
				logger.Error("FATAL: Failed to initialize FOCAS2 library", "error", err)
				return err // Это остановит запуск приложения
			}
			logger.Info("FOCAS2 library initialized successfully.")
			return nil
		},
	})
}

// InvokeRestoreConnections восстанавливает подключения и опросы при старте.
func InvokeRestoreConnections(lc fx.Lifecycle, fanucSvc interfaces.Usecases, dbRepo interfaces.FanucMachineRepository, logger *logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Restoring connections from the database...")
			machines, err := dbRepo.GetAll()
			if err != nil {
				logger.Error("Failed to get machine list from DB", "error", err)
				return nil // Не фатально, просто продолжаем
			}

			if len(machines) == 0 {
				logger.Info("No saved connections found to restore.")
				return nil
			}

			for _, machine := range machines {
				logger.Info("Attempting to restore connection", "sessionID", machine.SessionID, "endpoint", machine.EndpointURL)

				connInfo, _ := fanucSvc.RestoreConnection(machine)

				if connInfo.IsHealthy {
					logger.Info("Connection restored successfully in pool", "sessionID", machine.SessionID)
				} else {
					logger.Warn("Connection restored in pool but is unhealthy.", "sessionID", machine.SessionID)
				}

				if machine.Status == entities.StatusPolled && machine.Interval > 0 {
					interval := time.Duration(machine.Interval) * time.Millisecond
					logger.Info("Starting restored polling", "sessionID", machine.SessionID, "interval", interval)
					if err := fanucSvc.StartPolling(connInfo.SessionID, interval); err != nil {
						logger.Warn("Failed to start polling for restored session (it may be unhealthy)", "sessionID", machine.SessionID, "error", err)
					}
				}
			}
			return nil
		},
	})
}

// InvokeHttpServer запускает HTTP-сервер.
func InvokeHttpServer(lc fx.Lifecycle, cfg *config.AppConfig, h http.Handler, logger *logging.Logger) {
	serverAddr := ":" + cfg.ServerPort
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("HTTP Server is starting", "address", serverAddr)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Failed to start server", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server...")
			return server.Shutdown(ctx)
		},
	})
}
