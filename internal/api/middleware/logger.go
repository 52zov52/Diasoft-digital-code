package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func JSONLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		
		// маскируем PII в логах
		maskedURI := r.URL.Path
		if strings.Contains(maskedURI, "fio") || strings.Contains(maskedURI, "passport") {
			maskedURI = "***PII_MASKED***"
		}

		next.ServeHTTP(rw, r)

		slog.InfoContext(r.Context(), "HTTP Request",
			slog.String("method", r.Method),
			slog.String("uri", maskedURI),
			slog.Int("status", rw.status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", clientIP(r)),
		)
	})
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}