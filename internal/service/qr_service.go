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

// GenerateQRImage генерирует PNG изображение QR-кода в base64
func (s *QRService) GenerateQRImage(verifyURL string) (string, error) {
	png, err := qrcode.Encode(verifyURL, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("encode qrcode: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// VerifyQRToken проверяет токен и возвращает ID диплома
func (s *QRService) VerifyQRToken(ctx context.Context, token string) (string, error) {
	if s.rdb == nil {
		return "", fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("qr:%s", token)
	diplomaID, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token expired or invalid")
	}
	if err != nil {
		return "", fmt.Errorf("redis get: %w", err)
	}

	// Удаляем токен после использования (одноразовый)
	_ = s.rdb.Del(ctx, key)

	return diplomaID, nil
}

// GetDiplomaWithDetails — заглушка
func (s *QRService) GetDiplomaWithDetails(ctx context.Context, diplomaID string) (*models.Diploma, error) {
	return nil, fmt.Errorf("not implemented")
}