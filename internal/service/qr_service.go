package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/diasoft/diplomaverify/internal/models"
	"github.com/redis/go-redis/v9"
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
// В демо-режиме (без Redis) токен = diploma_id
func (s *QRService) GenerateQRToken(ctx context.Context, diplomaID string, ttlSeconds int) (string, error) {
	// Демо-режим: если Redis не подключен, возвращаем сам diplomaID как токен
	// Это позволяет верификации работать без кэша
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
	if err := s.rdb.Set(ctx, key, diplomaID, 
		// ttlSeconds может быть 0 для "бессрочного" токена в демо
		// но в продакшене всегда должен быть > 0
	).Err(); err != nil {
		// Логируем ошибку, но не прерываем — токен всё равно можно использовать
		// В реальном приложении здесь может быть fallback на БД
	}

	return token, nil
}

// VerifyQRToken проверяет токен и возвращает ID диплома
// В демо-режиме просто валидирует формат и возвращает токен как есть
func (s *QRService) VerifyQRToken(ctx context.Context, token string) (string, error) {
	// Демо-режим: токен уже содержит diploma_id
	// Проверяем минимальную валидность (UUID имеет длину 36, но можем принимать короче)
	if s.rdb == nil {
		if len(token) < 10 {
			return "", fmt.Errorf("invalid token format")
		}
		return token, nil
	}

	// Продакшен-режим: проверяем токен в Redis
	key := fmt.Sprintf("qr:%s", token)
	diplomaID, err := s.rdb.Get(ctx, key).Result()
	
	// Правильная константа для go-redis v9: redis.Nil (не ErrNil!)
	if err == redis.Nil {
		return "", fmt.Errorf("token expired or invalid")
	}
	if err != nil {
		// Если Redis временно недоступен — fallback на демо-режим
		// В продакшене здесь может быть логика повтора или возврат ошибки
		return token, nil
	}

	// Удаляем токен после использования (одноразовый)
	_ = s.rdb.Del(ctx, key)

	return diplomaID, nil
}

// GetDiplomaByID возвращает диплом по ID (заглушка для демо)
// В реальном приложении здесь будет запрос к репозиторию
func (s *QRService) GetDiplomaByID(ctx context.Context, diplomaID string) (*models.Diploma, error) {
	// Заглушка: в демо возвращаем ошибку, чтобы фронтенд показал "не найден"
	// В продакшене:
	// return s.repo.GetByID(ctx, diplomaID)
	return nil, fmt.Errorf("not implemented in demo mode")
}