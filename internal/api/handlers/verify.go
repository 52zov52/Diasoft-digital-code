package handlers

import (
	"net/http"

	"github.com/diasoft/diplomaverify/internal/service"
)

type VerifyHandler struct {
	diplomaSvc    *service.DiplomaService
	qrSvc         *service.QRService
	revocationSvc *service.RevocationService
}

func NewVerifyHandler(
	diplomaSvc *service.DiplomaService,
	qrSvc *service.QRService,
	revSvc *service.RevocationService,
) *VerifyHandler {
	return &VerifyHandler{
		diplomaSvc:    diplomaSvc,
		qrSvc:         qrSvc,
		revocationSvc: revSvc,
	}
}

// VerifyManual — ручная верификация по номеру диплома и коду вуза
func (h *VerifyHandler) VerifyManual(w http.ResponseWriter, r *http.Request) {
	number := r.URL.Query().Get("number")
	uniCode := r.URL.Query().Get("university_code")

	if number == "" || uniCode == "" {
		writeJSONError(w, http.StatusBadRequest, "missing parameters", "number and university_code required")
		return
	}

	dip, err := h.diplomaSvc.VerifyBySerial(r.Context(), number, uniCode)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "diploma not found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     dip.Status,
		"university": dip.UniversityID,
		"specialty":  dip.Specialty,
		"issue_year": dip.IssueYear,
		"fio_masked": maskPII(dip.FIOFull),
	})
}

// VerifyQR — верификация по токену из QR-кода
func (h *VerifyHandler) VerifyQR(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		writeJSONError(w, http.StatusBadRequest, "missing token", "")
		return
	}

	diplomaID, err := h.qrSvc.VerifyQRToken(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusGone, map[string]string{
			"status":  "expired",
			"message": "QR token has expired",
		})
		return
	}

	isRevoked, reason := h.revocationSvc.IsRevoked(r.Context(), diplomaID)
	if isRevoked {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"status":  "revoked",
			"reason":  reason,
			"message": "Диплом аннулирован",
		})
		return
	}

	dip, err := h.diplomaSvc.GetByID(r.Context(), diplomaID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"status": "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "valid",
		"university": dip.UniversityID,
		"specialty":  dip.Specialty,
		"issue_year": dip.IssueYear,
		"fio_masked": maskPII(dip.FIOFull),
	})
}

// BulkVerify — массовая верификация (заглушка)
func (h *VerifyHandler) BulkVerify(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "Bulk verification queued",
	})
}

// maskPII маскирует ФИО для безопасности
func maskPII(fio string) string {
	if len(fio) < 3 {
		return "***"
	}
	return fio[:2] + "***"
}