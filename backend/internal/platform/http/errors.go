package httphandler

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Error: code, Message: message})
}

func TenantNotFound(w http.ResponseWriter) {
	WriteError(w, http.StatusNotFound, "tenant_not_found", "unknown host or slug")
}

func ValidationError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "validation_error", message)
}
