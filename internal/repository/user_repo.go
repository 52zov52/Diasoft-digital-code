package repository

import (
"context"
"crypto/sha256"
"encoding/hex"
"errors"
"fmt"
"time"

"github.com/diasoft/diplomaverify/internal/models"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
"golang.org/x/crypto/bcrypt"
)

var (
ErrUserNotFound       = errors.New("user not found")
ErrInvalidCredentials = errors.New("invalid login or password")
ErrUserAlreadyExists  = errors.New("user with this login already exists")
)

type UserRepository interface {
Create(ctx context.Context, dto models.RegisterDTO, passwordHash string) (models.User, error)
GetByLogin(ctx context.Context, login string) (models.User, error)
UpdateLastLogin(ctx context.Context, userID string) error
GetUniversityByCode(ctx context.Context, code string) (models.University, error)
}

type userRepo struct {
pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
return &userRepo{pool: pool}
}

// HashPassword генерирует хеш пароля используя bcrypt
func HashPassword(password string) (string, error) {
bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
return "", fmt.Errorf("hash password: %w", err)
}
return string(bytes), nil
}

// CheckPasswordHash проверяет пароль против хеша
func CheckPasswordHash(password, hash string) bool {
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
return err == nil
}

func (r *userRepo) Create(ctx context.Context, dto models.RegisterDTO, passwordHash string) (models.User, error) {
var universityID *string

// Если роль university, проверяем существование вуза по коду
if dto.Role == models.RoleUniversity && dto.UniversityCode != "" {
uni, err := r.GetUniversityByCode(ctx, dto.UniversityCode)
if err != nil {
return models.User{}, fmt.Errorf("university not found: %w", err)
}
universityID = &uni.ID
}

query := `
INSERT INTO users (login, password_hash, role, university_id, created_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, login, role, university_id, created_at`

var u models.User
var uniID *string

err := r.pool.QueryRow(ctx, query, 
dto.Login, passwordHash, dto.Role, universityID).
Scan(&u.ID, &u.Login, &u.Role, &uniID, &u.CreatedAt)

if err != nil {
// Проверка на уникальный constraint
if pgErr := (*pgx.PgError)(nil); errors.As(err, &pgErr) {
if pgErr.Code() == "23505" { // unique_violation
return models.User{}, ErrUserAlreadyExists
}
}
return models.User{}, fmt.Errorf("query insert: %w", err)
}

u.PasswordHash = "" // Не возвращаем хеш
if uniID != nil {
u.UniversityID = *uniID
u.UniversityCode = dto.UniversityCode
}

return u, nil
}

func (r *userRepo) GetByLogin(ctx context.Context, login string) (models.User, error) {
query := `
SELECT id, login, password_hash, role, university_id, created_at, last_login_at
FROM users
WHERE login = $1`

var u models.User
var uniID *string
var lastLogin *time.Time

err := r.pool.QueryRow(ctx, query, login).Scan(
&u.ID, &u.Login, &u.PasswordHash, &u.Role, &uniID, &u.CreatedAt, &lastLogin)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return models.User{}, ErrUserNotFound
}
return models.User{}, fmt.Errorf("query select: %w", err)
}

u.LastLoginAt = lastLogin
if uniID != nil {
u.UniversityID = *uniID
}

// Получаем код вуза если есть university_id
if u.UniversityID != "" {
uni, err := r.getUniversityByID(ctx, u.UniversityID)
if err == nil {
u.UniversityCode = uni.Code
}
}

return u, nil
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, userID string) error {
query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
_, err := r.pool.Exec(ctx, query, userID)
if err != nil {
return fmt.Errorf("update last login: %w", err)
}
return nil
}

func (r *userRepo) GetUniversityByCode(ctx context.Context, code string) (models.University, error) {
query := `
SELECT id, code, official_name, public_key_pem, is_active, created_at, updated_at
FROM universities
WHERE code = $1 AND is_active = TRUE`

var uni models.University
err := r.pool.QueryRow(ctx, query, code).Scan(
&uni.ID, &uni.Code, &uni.OfficialName, &uni.PublicKeyPEM, 
&uni.IsActive, &uni.CreatedAt, &uni.UpdatedAt)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return models.University{}, errors.New("university not found or inactive")
}
return models.University{}, fmt.Errorf("query university: %w", err)
}

return uni, nil
}

func (r *userRepo) getUniversityByID(ctx context.Context, id string) (models.University, error) {
query := `
SELECT id, code, official_name, public_key_pem, is_active, created_at, updated_at
FROM universities
WHERE id = $1`

var uni models.University
err := r.pool.QueryRow(ctx, query, id).Scan(
&uni.ID, &uni.Code, &uni.OfficialName, &uni.PublicKeyPEM, 
&uni.IsActive, &uni.CreatedAt, &uni.UpdatedAt)

if err != nil {
return models.University{}, err
}

return uni, nil
}

// SessionRepository для управления сессиями
type SessionRepository interface {
CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
GetSession(ctx context.Context, tokenHash string) (bool, error)
RevokeSession(ctx context.Context, tokenHash string) error
CleanupExpiredSessions(ctx context.Context) error
}

type sessionRepo struct {
pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) SessionRepository {
return &sessionRepo{pool: pool}
}

func (r *sessionRepo) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
query := `
INSERT INTO user_sessions (user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, NOW())`

_, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt)
if err != nil {
return fmt.Errorf("create session: %w", err)
}
return nil
}

func (r *sessionRepo) GetSession(ctx context.Context, tokenHash string) (bool, error) {
query := `
SELECT EXISTS(
SELECT 1 FROM user_sessions 
WHERE token_hash = $1 AND expires_at > NOW() AND revoked_at IS NULL
)`

var exists bool
err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&exists)
if err != nil {
return false, fmt.Errorf("get session: %w", err)
}
return exists, nil
}

func (r *sessionRepo) RevokeSession(ctx context.Context, tokenHash string) error {
query := `UPDATE user_sessions SET revoked_at = NOW() WHERE token_hash = $1`
_, err := r.pool.Exec(ctx, query, tokenHash)
if err != nil {
return fmt.Errorf("revoke session: %w", err)
}
return nil
}

func (r *sessionRepo) CleanupExpiredSessions(ctx context.Context) error {
query := `DELETE FROM user_sessions WHERE expires_at < NOW()`
_, err := r.pool.Exec(ctx, query)
if err != nil {
return fmt.Errorf("cleanup sessions: %w", err)
}
return nil
}

// HashToken создает SHA-256 хеш токена
func HashToken(token string) string {
hash := sha256.Sum256([]byte(token))
return hex.EncodeToString(hash[:])
}
