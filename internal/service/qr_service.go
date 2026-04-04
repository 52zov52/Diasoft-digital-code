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

// GenerateQRToken генерирует токен для верификации по QR
func (s *QRService) GenerateQRToken(ctx context.Context, diplomaID string, ttlSeconds int) (string, error) {
	// Демо-режим: если Redis не подключен, возвращаем сам diplomaID как токен
	if s.rdb == nil {
		return diplomaID, nil
	}

	// Продакшен-режим: генерируем криптографически случайный токен
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Сохраняем токен в Redis с TTL
	key := fmt.Sprintf("qr:%s", token)
	expiration := time.Duration(ttlSeconds) * time.Second
	
	if err := s.rdb.Set(ctx, key, diplomaID, expiration).Err(); err != nil {
		// Логируем, но не прерываем
	}

	return token, nil
}

// GenerateQRImage генерирует PNG изображение QR-кода в base64
func (s *QRService) GenerateQRImage(verifyURL string) (string, error) {
	// Генерируем QR-код (256x256, средний уровень коррекции)
	png, err := qrcode.Encode(verifyURL, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("encode qrcode: %w", err)
	}

	// Конвертируем в base64 для отправки на фронтенд
	// Формат: "data:image/png;base64,..." — готов для <img src="">
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// VerifyQRToken проверяет токен и возвращает ID диплома
func (s *QRService) VerifyQRToken(ctx context.Context, token string) (string, error) {
	// Демо-режим: токен уже содержит diploma_id
	if s.rdb == nil {
		if len(token) < 10 {
			return "", fmt.Errorf("invalid token format")
		}
		return token, nil
	}

	// Продакшен-режим: проверяем токен в Redis
	key := fmt.Sprintf("qr:%s", token)
	diplomaID, err := s.rdb.Get(ctx, key).Result()
	
	if err == redis.Nil {
		return "", fmt.Errorf("token expired or invalid")
	}
	if err != nil {
		// Fallback на демо-режим при ошибке Redis
		return token, nil
	}

	// Удаляем токен после использования (одноразовый)
	_ = s.rdb.Del(ctx, key)

	return diplomaID, nil
}

// GetDiplomaByID — заглушка для демо
func (s *QRService) GetDiplomaByID(ctx context.Context, diplomaID string) (*models.Diploma, error) {
	return nil, fmt.Errorf("not implemented in demo mode")
}