package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/diasoft/diplomaverify/internal/api/middleware"
	"github.com/diasoft/diplomaverify/internal/service"
)

type QRHandler struct {
	svc     *service.QRService
	baseURL string
}

func NewQRHandler(svc *service.QRService) *QRHandler {
	return &QRHandler{
		svc:     svc,
		baseURL: "https://diploma-verify-backend.onrender.com/api/v1",
	}
}

// Generate генерирует QR-код для диплома
func (h *QRHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DiplomaID string `json:"diploma_id" validate:"required"`
		TTL       int    `json:"ttl" validate:"required,min=300,max=604800"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	if errs := middleware.ValidateStruct(req); errs != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "validation failed", errs)
		return
	}

	// Генерируем токен
	token, err := h.svc.GenerateQRToken(r.Context(), req.DiplomaID, req.TTL)
	if err != nil {
		slog.Error("Failed to generate QR token", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token", "")
		return
	}

	// Формируем ссылку для верификации
	verifyURL := fmt.Sprintf("%s/verify/qr/%s", h.baseURL, token)

	// Генерируем QR-изображение
	qrImage, err := h.svc.GenerateQRImage(verifyURL)
	if err != nil {
		slog.Error("Failed to generate QR image", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to generate image", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token":           token,
		"qr_image_base64": qrImage,
		"verify_url":      verifyURL,
	})
}