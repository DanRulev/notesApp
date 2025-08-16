package app

import (
	"context"
	"log"
	"noteApp/internal/config"
	"noteApp/internal/handler"
	"noteApp/internal/repository"
	"noteApp/internal/server"
	"noteApp/internal/service"
	"noteApp/pkg/db"
	"noteApp/pkg/hasher"
	"noteApp/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func Start() {
	log.Println("booting application...")

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	zapLogger, cleanUp, err := logger.Init(cfg.Logger)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer cleanUp()

	zapLogger.Info("application boot started",
		zap.String("service", "noteApp"),
		zap.Bool("development", cfg.Logger.Development),
	)

	zapLogger.Info("config loaded",
		zap.String("server_port", cfg.Server.Port),
		zap.Duration("shutdown_timeout", cfg.Server.ShutdownTimeout),
		zap.Duration("access_ttl", cfg.Auth.AccessTokenTTL),
		zap.Duration("refresh_ttl", cfg.Auth.RefreshTokenTTL),
		zap.String("db_host", cfg.DB.Host),
		zap.String("db_port", cfg.DB.Port),
	)

	zapLogger.Info("connecting to database...")
	dbConn, closeDB, err := db.New(cfg.DB)
	if err != nil {
		zapLogger.Fatal("failed to connect to database",
			zap.Error(err),
		)
	}
	defer func() {
		if err := closeDB(); err != nil {
			zapLogger.Error("failed to close database connection",
				zap.Error(err),
			)
		} else {
			zapLogger.Info("database connection closed")
		}
	}()

	zapLogger.Info("database connected successfully",
		zap.String("database", cfg.DB.Name),
	)

	zapLogger.Info("initializing repositories")
	repos := repository.NewRepository(dbConn.DB, zapLogger)

	zapLogger.Info("initializing password hasher")
	hasher := hasher.NewHasher()

	zapLogger.Info("initializing services")
	services := service.NewService(repos, hasher, cfg.Auth, zapLogger)

	zapLogger.Info("initializing HTTP handlers")
	handlers := handler.NewHandler(services, zapLogger, cfg.Auth.RefreshTokenTTL)

	zapLogger.Info("initializing HTTP server",
		zap.String("port", cfg.Server.Port),
		zap.Duration("read_timeout", cfg.Server.ReadTimeout),
		zap.Duration("write_timeout", cfg.Server.WriteTimeout),
	)
	server := server.NewServer(cfg.Server, handlers.Init())

	zapLogger.Info("starting HTTP server", zap.String("port", cfg.Server.Port))
	go func() {
		if err := server.Run(); err != nil {
			zapLogger.Fatal("failed to run server",
				zap.Error(err),
				zap.String("port", cfg.Server.Port),
			)
		}
	}()

	zapLogger.Info("server is running",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port),
	)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-exit

	zapLogger.Info("shutdown signal received",
		zap.String("signal", sig.String()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	zapLogger.Info("gracefully shutting down server...")
	if err := server.Stop(ctx); err != nil {
		zapLogger.Fatal("failed to shutdown server gracefully",
			zap.Error(err),
		)
	}

	zapLogger.Info("server stopped gracefully. Goodbye!")
}
