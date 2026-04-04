package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidatePayload(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next(w, r)
			return
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields() // защита от неизвестных полей

		var payload any
		if err := dec.Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json payload", err.Error())
			return
		}

		// Валидация структур будет применяться в хендлерах через теги
		// Здесь только базовая проверка JSON-формата
		next(w, r)
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   msg,
		"details": details,
	})
}

// ValidateStruct - вспомогательная функция для хендлеров
func ValidateStruct(v any) []map[string]string {
	errs := validate.Struct(v)
	if errs == nil {
		return nil
	}

	var out []map[string]string
	for _, err := range errs.(validator.ValidationErrors) {
		out = append(out, map[string]string{
			"field":   err.Field(),
			"tag":     err.Tag(),
			"message": err.Error(),
		})
	}
	return out
}