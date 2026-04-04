package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/diasoft/diplomaverify/internal/models"
	"github.com/diasoft/diplomaverify/internal/repository"
)

type DiplomaService struct {
	repo repository.DiplomaRepository
}

func NewDiplomaService(repo repository.DiplomaRepository) *DiplomaService {
	return &DiplomaService{repo: repo}
}

func (s *DiplomaService) CreateDiploma(ctx context.Context, dto models.CreateDiplomaDTO) (models.Diploma, error) {
	// Генерация хеша канонической записи (упрощенно для демо, в проде используется canonical JSON)
	payload := fmt.Sprintf("%s|%d|%s|%s", dto.SerialNumber, dto.IssueYear, dto.Specialty, dto.FIOFull)
	hash := sha256.Sum256([]byte(payload))
	hashHex := hex.EncodeToString(hash[:])
	
	// Заглушка подписи вуза (в Фазе 4 будет реальная Ed25519)
	signature := "stub_ed25519_signature_" + hashHex[:16]

	d, err := s.repo.Create(ctx, dto, hashHex, signature)
	if err != nil {
		return models.Diploma{}, fmt.Errorf("service create: %w", err)
	}
	return d, nil
}

func (s *DiplomaService) VerifyBySerial(ctx context.Context, serial, uniCode string) (models.Diploma, error) {
	d, err := s.repo.GetBySerialAndUniCode(ctx, serial, uniCode)
	if err != nil {
		return models.Diploma{}, fmt.Errorf("service verify: %w", err)
	}
	return d, nil
}

func (s *DiplomaService) RevokeDiploma(ctx context.Context, id, reason string) error {
	return s.repo.Revoke(ctx, id, reason)
}

// GetByID возвращает диплом по его ID (используется при верификации по QR)
func (s *DiplomaService) GetByID(ctx context.Context, id string) (models.Diploma, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Diploma{}, fmt.Errorf("service get by id: %w", err)
	}
	return d, nil
}