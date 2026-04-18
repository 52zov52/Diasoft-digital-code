package service

import (
"context"
"errors"
"fmt"
"time"

"github.com/diasoft/diplomaverify/internal/models"
"github.com/diasoft/diplomaverify/internal/repository"
"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
userRepo    repository.UserRepository
sessionRepo repository.SessionRepository
jwtSecret   []byte
tokenTTL    time.Duration
}

func NewAuthService(
userRepo repository.UserRepository,
sessionRepo repository.SessionRepository,
jwtSecret string,
tokenTTL time.Duration,
) *AuthService {
return &AuthService{
userRepo:    userRepo,
sessionRepo: sessionRepo,
jwtSecret:   []byte(jwtSecret),
tokenTTL:    tokenTTL,
}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, dto models.RegisterDTO) (models.User, error) {
// Валидация роли и university_code
if dto.Role == models.RoleUniversity && dto.UniversityCode == "" {
return models.User{}, errors.New("university code is required for university role")
}

// Хэширование пароля
passwordHash, err := repository.HashPassword(dto.Password)
if err != nil {
return models.User{}, fmt.Errorf("hash password: %w", err)
}

// Создание пользователя
user, err := s.userRepo.Create(ctx, dto, passwordHash)
if err != nil {
return models.User{}, err
}

return user, nil
}

// Login аутентифицирует пользователя и возвращает JWT токен
func (s *AuthService) Login(ctx context.Context, dto models.LoginDTO) (models.AuthResponse, error) {
// Поиск пользователя по логину
user, err := s.userRepo.GetByLogin(ctx, dto.Login)
if err != nil {
if errors.Is(err, repository.ErrUserNotFound) {
return models.AuthResponse{}, repository.ErrInvalidCredentials
}
return models.AuthResponse{}, err
}

// Проверка пароля
if !repository.CheckPasswordHash(dto.Password, user.PasswordHash) {
return models.AuthResponse{}, repository.ErrInvalidCredentials
}

// Генерация JWT токена
expiresAt := time.Now().Add(s.tokenTTL)
token, err := s.generateJWT(user, expiresAt)
if err != nil {
return models.AuthResponse{}, fmt.Errorf("generate token: %w", err)
}

// Сохранение сессии
tokenHash := repository.HashToken(token)
err = s.sessionRepo.CreateSession(ctx, user.ID, tokenHash, expiresAt)
if err != nil {
return models.AuthResponse{}, fmt.Errorf("create session: %w", err)
}

// Обновление last_login_at
_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

// Очистка чувствительных данных
user.PasswordHash = ""

return models.AuthResponse{
User:      user,
Token:     token,
ExpiresAt: expiresAt.Unix(),
}, nil
}

// Logout завершает сессию пользователя
func (s *AuthService) Logout(ctx context.Context, token string) error {
tokenHash := repository.HashToken(token)
return s.sessionRepo.RevokeSession(ctx, tokenHash)
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (models.User, error) {
// Проверка сессии
tokenHash := repository.HashToken(tokenString)
exists, err := s.sessionRepo.GetSession(ctx, tokenHash)
if err != nil || !exists {
return models.User{}, errors.New("invalid or expired session")
}

// Парсинг токена
claims, err := s.parseJWT(tokenString)
if err != nil {
return models.User{}, err
}

// Получение пользователя
user, err := s.userRepo.GetByLogin(ctx, claims.Login)
if err != nil {
return models.User{}, err
}

user.PasswordHash = ""
return user, nil
}

// Claims для JWT токена
type UserClaims struct {
UserID string           `json:"user_id"`
Login  string           `json:"login"`
Role   models.UserRole  `json:"role"`
jwt.RegisteredClaims
}

func (s *AuthService) generateJWT(user models.User, expiresAt time.Time) (string, error) {
claims := UserClaims{
UserID: user.ID,
Login:  user.Login,
Role:   user.Role,
RegisteredClaims: jwt.RegisteredClaims{
ExpiresAt: jwt.NewNumericDate(expiresAt),
IssuedAt:  jwt.NewNumericDate(time.Now()),
Issuer:    "diplomaverify",
},
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString(s.jwtSecret)
if err != nil {
return "", fmt.Errorf("sign token: %w", err)
}

return tokenString, nil
}

func (s *AuthService) parseJWT(tokenString string) (*UserClaims, error) {
token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
}
return s.jwtSecret, nil
})

if err != nil {
return nil, fmt.Errorf("parse token: %w", err)
}

claims, ok := token.Claims.(*UserClaims)
if !ok || !token.Valid {
return nil, errors.New("invalid token claims")
}

return claims, nil
}
