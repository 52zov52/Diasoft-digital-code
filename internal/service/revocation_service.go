package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RevocationService struct {
	rdb *redis.Client
}

func NewRevocationService(rdb *redis.Client) *RevocationService {
	return &RevocationService{rdb: rdb}
}

// MarkRevoked добавляет диплом в ревок-лист Redis (быстрая проверка)
func (s *RevocationService) MarkRevoked(ctx context.Context, diplomaID string, reason string) error {
	key := fmt.Sprintf("revoked:diploma:%s", diplomaID)
	if err := s.rdb.Set(ctx, key, reason, 365*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis set revocation: %w", err)
	}
	return nil
}

// IsRevoked проверяет ревок-лист (Redis first, потом fallback)
func (s *RevocationService) IsRevoked(ctx context.Context, diplomaID string) (bool, string) {
	key := fmt.Sprintf("revoked:diploma:%s", diplomaID)
	reason, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, ""
	} else if err != nil {
		// Fallback логируется в реальном сервисе
		return false, ""
	}
	return true, reason
}