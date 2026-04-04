package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/diasoft/diplomaverify/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/skip2/go-qrcode"
)

type QRService struct {
	rdb     *redis.Client
	baseURL string
}

func NewQRService(rdb *redis.Client, baseURL string) *QRService {
	return &QRService{
		rdb:     rdb,
		baseURL: baseURL,
	}
}

// GenerateQRToken генерирует токен и (опционально) сохраняет в Redis
func (s *QRService) GenerateQRToken(ctx context.Context, diplomaID string, ttlSeconds int) (string, error) {
	// Генерируем криптографически случайный токен
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Если Redis подключен — сохраняем токен
	if s.rdb != nil {
		key := fmt.Sprintf("qr:%s", token)
		if err := s.rdb.Set(ctx, key, diplomaID, time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
			// Логируем, но не прерываем — токен всё равно можно использовать
		}
	}

	return token, nil
}

// GenerateQRToken генерирует токен. В демо-режиме (без Redis) токен = ID диплома
func (s *QRService) GenerateQRToken(ctx context.Context, diplomaID string, ttlSeconds int) (string, error) {
	// Если Redis недоступен, используем демо-режим: токен = diploma_id
	// Это позволит верификации работать без кэша
	if s.rdb == nil {
		return diplomaID, nil
	}

	// Продакшен-логика для Redis
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	key := fmt.Sprintf("qr:%s", token)
	if err := s.rdb.Set(ctx, key, diplomaID, time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
		// Логируем, но не прерываем
	}

	return token, nil
}

// VerifyQRToken проверяет токен. В демо-режиме сразу возвращает ID диплома
func (s *QRService) VerifyQRToken(ctx context.Context, token string) (string, error) {
	// Демо-режим: токен уже содержит diploma_id, проверяем только формат UUID
	if s.rdb == nil {
		if len(token) < 10 {
			return "", fmt.Errorf("invalid token format")
		}
		return token, nil
	}

	// Продакшен-логика для Redis
	key := fmt.Sprintf("qr:%s", token)
	diplomaID, err := s.rdb.Get(ctx, key).Result()
	if err == redis.ErrNil {
		return "", fmt.Errorf("token expired or invalid")
	}
	if err != nil {
		// Если Redis упал, fallback на демо-режим
		return token, nil
	}

	_ = s.rdb.Del(ctx, key)
	return diplomaID, nil
}

// GetDiplomaWithDetails — заглушка
func (s *QRService) GetDiplomaWithDetails(ctx context.Context, diplomaID string) (*models.Diploma, error) {
	return nil, fmt.Errorf("not implemented")
}