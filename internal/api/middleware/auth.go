package middleware

import (
"context"
"net/http"
"strings"

"github.com/diasoft/diplomaverify/internal/models"
"github.com/diasoft/diplomaverify/internal/service"
)

// contextKey для хранения данных пользователя в контексте
type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware проверяет JWT токен и добавляет пользователя в контекст
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
return
}

// Убираем префикс "Bearer "
token := strings.TrimPrefix(authHeader, "Bearer ")
if token == authHeader {
http.Error(w, `{"error": "invalid authorization format"}`, http.StatusUnauthorized)
return
}

// Валидация токена
user, err := authService.ValidateToken(r.Context(), token)
if err != nil {
http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnauthorized)
return
}

// Добавляем пользователя в контекст
ctx := context.WithValue(r.Context(), UserContextKey, user)
next.ServeHTTP(w, r.WithContext(ctx))
})
}
}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(ctx context.Context) (models.User, bool) {
user, ok := ctx.Value(UserContextKey).(models.User)
return user, ok
}

// RequireRole middleware проверяет роль пользователя
func RequireRole(allowedRoles ...models.UserRole) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
user, ok := GetUserFromContext(r.Context())
if !ok {
http.Error(w, `{"error": "user not found in context"}`, http.StatusInternalServerError)
return
}

// Проверка роли
allowed := false
for _, role := range allowedRoles {
if user.Role == role {
allowed = true
break
}
}

if !allowed {
http.Error(w, `{"error": "insufficient permissions"}`, http.StatusForbidden)
return
}

next.ServeHTTP(w, r)
})
}
}
