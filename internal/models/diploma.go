package models

import "time"

type Diploma struct {
	ID               string
	UniversityID     string
	SerialNumber     string
	Specialty        string
	IssueYear        int
	FIOFull          string // Расшифровывается при чтении
	PassportSN       string // Опционально, расшифровывается при чтении
	DocumentHash     string
	DigitalSignature string
	Status           string
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

type CreateDiplomaDTO struct {
	UniversityID string `json:"university_id"`  // ← Просто поле, без валидации
	SerialNumber string `json:"serial" validate:"required"`
	FIOFull      string `json:"fio" validate:"required"`
	IssueYear    int    `json:"year" validate:"required,min=1900,max=2100"`
	Specialty    string `json:"specialty" validate:"required"`
	PassportSN   string `json:"passport_sn,omitempty"`
}