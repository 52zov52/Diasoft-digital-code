package api

import (
	"net/http"
	"time"

	"github.com/diasoft/diplomaverify/internal/api/handlers"
	"github.com/diasoft/diplomaverify/internal/api/middleware"
	"github.com/diasoft/diplomaverify/internal/config"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(
	cfg *config.Config,
	dipH *handlers.DiplomaHandler,
	qrH *handlers.QRHandler,
	verifyH *handlers.VerifyHandler,
) http.Handler {
	r := chi.NewRouter()

	// 1. Глобальные middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.JSONLogger)
	r.Use(chimiddleware.Timeout(10 * time.Second))
	r.Use(middleware.ContextWithTimeout(cfg.Timeout))

	// 2. CORS Middleware (ОБЯЗАТЕЛЬНО ДО всех маршрутов!)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// 3. Публичные системные эндпоинты (без префикса /api)
	r.Get("/health", handlers.HealthCheck)
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/api/openapi.yaml")
	})

	// 4. API v1 — ВСЕ маршруты с префиксом /api/v1
	r.Route("/api/v1", func(r chi.Router) {
		
		// === ПУБЛИЧНЫЕ МАРШРУТЫ (верификация) ===
		// Важно: сначала конкретные пути, потом с параметрами
		r.Get("/verify", verifyH.VerifyManual)              // GET /api/v1/verify
		r.Get("/verify/qr/{token}", verifyH.VerifyQR)       // GET /api/v1/verify/qr/{token}
		r.Post("/diplomas/bulk-verify", verifyH.BulkVerify) // POST /api/v1/diplomas/bulk-verify

		// === МАРШРУТЫ ВУЗА (требуют авторизации — пока отключено для тестов) ===
		r.Post("/diplomas/bulk", dipH.BulkUpload)           // POST /api/v1/diplomas/bulk
		r.Post("/diplomas/{id}/revoke", dipH.Revoke)        // POST /api/v1/diplomas/{id}/revoke
		r.Post("/qr/generate", qrH.Generate)                // POST /api/v1/qr/generate
	})

	return r
}