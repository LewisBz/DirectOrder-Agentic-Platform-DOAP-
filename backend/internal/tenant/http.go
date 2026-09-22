package tenant

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
)

func GetCurrent(res Resolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeTenant(w, r, res, platmw.HostFromContext(r.Context()), "")
	}
}

func GetBySlug(res Resolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeTenant(w, r, res, platmw.HostFromContext(r.Context()), chi.URLParam(r, "slug"))
	}
}

func writeTenant(w http.ResponseWriter, r *http.Request, res Resolver, host, slug string) {
	w.Header().Set("Content-Type", "application/json")
	if res == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "tenant_not_found", "message": "unknown host or slug"})
		return
	}
	t, err := res.Resolve(r.Context(), host, slug)
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "tenant_not_found", "message": "unknown host or slug"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "validation_error", "message": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(t)
}
