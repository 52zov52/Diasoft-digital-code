package handlers

import (
"encoding/json"
"errors"
"net/http"

"github.com/diasoft/diplomaverify/internal/models"
"github.com/diasoft/diplomaverify/internal/service"
)

type AuthHandler struct {
authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
return &AuthHandler{authService: authService}
}

// Register обрабатывает регистрацию нового пользователя
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
var dto models.RegisterDTO
if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
return
}

if err := validateStruct(dto); err != nil {
http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
return
}

user, err := h.authService.Register(r.Context(), dto)
if err != nil {
status := http.StatusInternalServerError
if err.Error() == "university code is required for university role" {
status = http.StatusBadRequest
} else if err.Error() == "university not found or inactive" {
status = http.StatusNotFound
} else if err.Error() == "user with this login already exists" {
status = http.StatusConflict
}
http.Error(w, `{"error": "`+err.Error()+`"}`, status)
return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(map[string]interface{}{
"message": "User registered successfully",
"user":    user,
})
}

// Login обрабатывает вход пользователя
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
var dto models.LoginDTO
if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
return
}

response, err := h.authService.Login(r.Context(), dto)
if err != nil {
status := http.StatusInternalServerError
if err.Error() == "invalid login or password" || err.Error() == "user not found" {
status = http.StatusUnauthorized
}
http.Error(w, `{"error": "`+err.Error()+`"}`, status)
return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(response)
}

// Logout обрабатывает выход пользователя
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
token := r.Header.Get("Authorization")
if token == "" {
http.Error(w, `{"error": "authorization header required"}`, http.StatusBadRequest)
return
}

if len(token) > 7 && token[:7] == "Bearer " {
token = token[7:]
}

if err := h.authService.Logout(r.Context(), token); err != nil {
http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{
"message": "Logged out successfully",
})
}

// Me возвращает информацию о текущем пользователе
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
token := r.Header.Get("Authorization")
if token == "" {
http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
return
}

if len(token) > 7 && token[:7] == "Bearer " {
token = token[7:]
}

user, err := h.authService.ValidateToken(r.Context(), token)
if err != nil {
http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnauthorized)
return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(user)
}

// GetUniversities возвращает список всех активных вузов (для формы регистрации)
// GET /api/v1/auth/universities
func (h *AuthHandler) GetUniversities(w http.ResponseWriter, r *http.Request) {
universities := []models.University{
{ID: "1", Code: "MSU", OfficialName: "Московский Государственный Университет", IsActive: true},
{ID: "2", Code: "SPBU", OfficialName: "Санкт-Петербургский Государственный Университет", IsActive: true},
{ID: "3", Code: "HSE", OfficialName: "Высшая Школа Экономики", IsActive: true},
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(universities)
}

func validateStruct(dto interface{}) error {
switch v := dto.(type) {
case models.RegisterDTO:
if v.Login == "" {
return errors.New("login is required")
}
if len(v.Login) < 3 {
return errors.New("login must be at least 3 characters")
}
if v.Password == "" {
return errors.New("password is required")
}
if len(v.Password) < 6 {
return errors.New("password must be at least 6 characters")
}
if v.Role == "" {
return errors.New("role is required")
}
}
return nil
}
