package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/diasoft/diplomaverify/internal/api"
	"github.com/diasoft/diplomaverify/internal/api/handlers"
	"github.com/diasoft/diplomaverify/internal/config"
	"github.com/diasoft/diplomaverify/internal/crypto"
	"github.com/diasoft/diplomaverify/internal/db"
	"github.com/diasoft/diplomaverify/internal/repository"
	"github.com/diasoft/diplomaverify/internal/service"
)

func main() {
	ctx := context.Background()

	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Настройка логгера
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.SlogLevel(),
	})))

	// 3. Инициализация криптографии
	cipher, err := crypto.NewAESCipher(cfg.AESKeyHex)
	if err != nil {
		slog.Error("Failed to init AES cipher", "error", err)
		os.Exit(1)
	}

	// 4. Подключение к PostgreSQL
	var pool *pgxpool.Pool
	dbDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	pool, err = db.NewPool(ctx, db.PostgresConfig{
		DSN:             dbDSN,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	})
	if err != nil {
		slog.Error("PostgreSQL connection failed", "error", err)
		os.Exit(1) // БД обязательна для работы
	}
	defer pool.Close()

	// 5. Подключение к Redis (опционально)
	var rdb *redis.Client
	if cfg.RedisURL != "" {
		rdb, err = db.NewRedisClient(ctx, db.RedisConfig{URL: cfg.RedisURL})
		if err != nil {
			slog.Warn("Redis connection failed, continuing without cache", "error", err)
		} else {
			defer rdb.Close()
		}
	}

	// 6. Создаём репозиторий ← ИСПРАВЛЕНО: используем = вместо :=
	var diplomaRepo repository.DiplomaRepository
	if pool != nil {
		diplomaRepo = repository.NewDiplomaRepository(pool, cipher) // ← = а не :=
	}

	// 7. Создаём сервисы ← ИСПРАВЛЕНО: проверяем rdb != nil
	diplomaSvc := service.NewDiplomaService(diplomaRepo)

	var qrSvc *service.QRService
	var revSvc *service.RevocationService
	if rdb != nil {
		qrSvc = service.NewQRService(rdb, "http://localhost:8080/api/v1")
		revSvc = service.NewRevocationService(rdb)
	} else {
		// Заглушки без Redis — для dev-режима
		qrSvc = service.NewQRService(nil, "http://localhost:8080/api/v1")
		revSvc = service.NewRevocationService(nil)
	}

	// 8. Создаём хендлеры
	diplomaHandler := handlers.NewDiplomaHandler(diplomaSvc)
	qrHandler := handlers.NewQRHandler(qrSvc)
	verifyHandler := handlers.NewVerifyHandler(diplomaSvc, qrSvc, revSvc)

	// 9. Создаём роутер
	router := api.NewRouter(cfg, diplomaHandler, qrHandler, verifyHandler)

	// 10. Запускаем HTTP-сервер
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// 11. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")

	shutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("Server forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}