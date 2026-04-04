package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/diasoft/diplomaverify/internal/api/middleware"
	"github.com/diasoft/diplomaverify/internal/models"
	"github.com/diasoft/diplomaverify/internal/service"
)

type DiplomaHandler struct {
	svc *service.DiplomaService
}

func NewDiplomaHandler(svc *service.DiplomaService) *DiplomaHandler {
	return &DiplomaHandler{svc: svc}
}

func (h *DiplomaHandler) BulkUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UniversityID string                `json:"university_id" validate:"required,uuid"`
		Records      []models.CreateDiplomaDTO `json:"records" validate:"required,dive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	// Логирование для отладки
	slog.Info("Received bulk upload request", 
		"university_id", req.UniversityID,
		"records_count", len(req.Records))

	if errs := middleware.ValidateStruct(req); errs != nil {
		slog.Error("Validation failed", "errors", errs)
		writeJSONError(w, http.StatusUnprocessableEntity, "validation failed", errs)
		return
	}

	createdCount := 0
	for _, rec := range req.Records {
		rec.UniversityID = req.UniversityID
		if _, err := h.svc.CreateDiploma(r.Context(), rec); err != nil {
			slog.Error("Failed to create diploma", "error", err, "serial", rec.SerialNumber)
			continue 
		}
		createdCount++
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":       "Bulk upload processed",
		"records_added": createdCount,
		"total_sent":    len(req.Records),
	})
}

func (h *DiplomaHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing id", "path parameter required")
		return
	}

	var req struct { Reason string `json:"reason"` }
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.RevokeDiploma(r.Context(), id, req.Reason); err != nil {
		writeJSONError(w, http.StatusNotFound, "revoke failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "diploma revoked"})
}
// GetRecent возвращает последние загруженные дипломы
func (h *DiplomaHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	universityID := r.URL.Query().Get("university_id")
	if universityID == "" {
		writeJSONError(w, http.StatusBadRequest, "university_id required", "")
		return
	}

	// Получаем последние 10 дипломов (заглушка — в продакшене нужен реальный запрос к БД)
	// Пока вернём пустой массив
	writeJSON(w, http.StatusOK, []map[string]interface{}{})
}