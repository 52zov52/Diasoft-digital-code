package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/diasoft/diplomaverify/internal/crypto"
	"github.com/diasoft/diplomaverify/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDiplomaNotFound = errors.New("diploma not found")

type DiplomaRepository interface {
	Create(ctx context.Context, dto models.CreateDiplomaDTO, hash, signature string) (models.Diploma, error)
	GetBySerialAndUniCode(ctx context.Context, serial, uniCode string) (models.Diploma, error)
	GetByID(ctx context.Context, id string) (models.Diploma, error)
	Revoke(ctx context.Context, id, reason string) error
}

type diplomaRepo struct {
	pool   *pgxpool.Pool
	cipher *crypto.AESCipher
}

func NewDiplomaRepository(pool *pgxpool.Pool, cipher *crypto.AESCipher) DiplomaRepository {
	return &diplomaRepo{pool: pool, cipher: cipher}
}

func (r *diplomaRepo) Create(ctx context.Context, dto models.CreateDiplomaDTO, hash, signature string) (models.Diploma, error) {
	fioEnc, err := r.cipher.Encrypt(dto.FIOFull)
	if err != nil {
		return models.Diploma{}, fmt.Errorf("encrypt fio: %w", err)
	}
	
	passEnc := ""
	if dto.PassportSN != "" {
		passEnc, err = r.cipher.Encrypt(dto.PassportSN)
		if err != nil {
			return models.Diploma{}, fmt.Errorf("encrypt passport: %w", err)
		}
	}

	query := `
		INSERT INTO diplomas (university_id, serial_number, specialty, issue_year, 
							  encrypted_fio_full, encrypted_passport_sn, document_hash, digital_signature, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, created_at`

	var d models.Diploma
	err = r.pool.QueryRow(ctx, query, 
		dto.UniversityID, dto.SerialNumber, dto.Specialty, dto.IssueYear,
		fioEnc, passEnc, hash, signature, "active").
		Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return models.Diploma{}, fmt.Errorf("query insert: %w", err)
	}

	d.SerialNumber = dto.SerialNumber
	d.Specialty = dto.Specialty
	d.IssueYear = dto.IssueYear
	d.FIOFull = dto.FIOFull
	d.DocumentHash = hash
	d.DigitalSignature = signature
	d.Status = "active"
	return d, nil
}

func (r *diplomaRepo) GetBySerialAndUniCode(ctx context.Context, serial, uniCode string) (models.Diploma, error) {
	query := `
		SELECT d.id, u.id, d.serial_number, d.specialty, d.issue_year, d.encrypted_fio_full, 
			   d.encrypted_passport_sn, d.document_hash, d.digital_signature, d.status, 
			   d.revoked_at, d.created_at
		FROM diplomas d JOIN universities u ON d.university_id = u.id
		WHERE d.serial_number = $1 AND u.code = $2`

	var d models.Diploma
	var fioEnc, passEnc string
	var revokedAt *time.Time

	err := r.pool.QueryRow(ctx, query, serial, uniCode).Scan(
		&d.ID, &d.UniversityID, &d.SerialNumber, &d.Specialty, &d.IssueYear,
		&fioEnc, &passEnc, &d.DocumentHash, &d.DigitalSignature, &d.Status,
		&revokedAt, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Diploma{}, ErrDiplomaNotFound
		}
		return models.Diploma{}, fmt.Errorf("query select: %w", err)
	}

	d.RevokedAt = revokedAt
	d.FIOFull, _ = r.cipher.Decrypt(fioEnc)
	if passEnc != "" {
		d.PassportSN, _ = r.cipher.Decrypt(passEnc)
	}
	return d, nil
}

func (r *diplomaRepo) GetByID(ctx context.Context, id string) (models.Diploma, error) {
	query := `SELECT id, university_id, serial_number, specialty, issue_year, encrypted_fio_full, 
			  encrypted_passport_sn, document_hash, digital_signature, status, revoked_at, created_at 
			  FROM diplomas WHERE id = $1`
	
	var d models.Diploma
	var fioEnc, passEnc string
	var revokedAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.UniversityID, &d.SerialNumber, &d.Specialty, &d.IssueYear,
		&fioEnc, &passEnc, &d.DocumentHash, &d.DigitalSignature, &d.Status,
		&revokedAt, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Diploma{}, ErrDiplomaNotFound
		}
		return models.Diploma{}, fmt.Errorf("query by id: %w", err)
	}

	d.RevokedAt = revokedAt
	d.FIOFull, _ = r.cipher.Decrypt(fioEnc)
	if passEnc != "" {
		d.PassportSN, _ = r.cipher.Decrypt(passEnc)
	}
	return d, nil
}

func (r *diplomaRepo) Revoke(ctx context.Context, id, reason string) error {
	now := time.Now().UTC()
	res, err := r.pool.Exec(ctx, 
		"UPDATE diplomas SET status = 'revoked', revoked_at = $1, revocation_reason = $2 WHERE id = $3 AND status = 'active'",
		now, reason, id)
	if err != nil {
		return fmt.Errorf("update revoke: %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("diploma already revoked or not found")
	}
	return nil
}