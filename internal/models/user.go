package models

import "time"

// UserRole определяет роль пользователя
type UserRole string

const (
	RoleStudent   UserRole = "student"
	RoleUniversity UserRole = "university"
	RoleHR        UserRole = "hr"
)

// User представляет пользователя системы
type User struct {
	ID              string     `json:"id"`
	Login           string     `json:"login"`            // Логин (уникальный)
	PasswordHash    string     `json:"-"`                // Хеш пароля (не возвращается в JSON)
	Role            UserRole   `json:"role"`             // Роль: student, university, hr
	UniversityCode  string     `json:"university_code,omitempty"` // Код вуза (только для роли university)
	UniversityID    string     `json:"university_id,omitempty"`   // ID вуза в системе (связь с таблицей universities)
	CreatedAt       time.Time  `json:"created_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

// RegisterDTO - данные для регистрации
type RegisterDTO struct {
	Login          string   `json:"login" validate:"required,min=3,max=50"`
	Password       string   `json:"password" validate:"required,min=6,max=100"`
	Role           UserRole `json:"role" validate:"required,oneof=student university hr"`
	UniversityCode string   `json:"university_code,omitempty"` // Требуется только для university
}

// LoginDTO - данные для входа
type LoginDTO struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse - ответ при успешной аутентификации
type AuthResponse struct {
	User      User   `json:"user"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}

// University represents a university in the system
type University struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`         // Уникальный код вуза (для CSV загрузки)
	OfficialName string    `json:"official_name"`
	PublicKeyPEM string    `json:"public_key_pem"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUniversityDTO - данные для создания вуза
type CreateUniversityDTO struct {
	Code         string `json:"code" validate:"required,min=3,max=16"`
	OfficialName string `json:"official_name" validate:"required"`
	PublicKeyPEM string `json:"public_key_pem" validate:"required"`
}
