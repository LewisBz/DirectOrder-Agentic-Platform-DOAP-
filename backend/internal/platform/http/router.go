package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/tenant"
)

func NewRouter(db *database.DB, tenants tenant.Resolver) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(corsAllow)
	r.Use(platmw.IgnoreClientTenantID)
	r.Get("/healthz", Healthz())
	r.Get("/readyz", Readyz(db))
	r.Get("/v1/tenants/current", tenant.GetCurrent(tenants))
	r.Get("/v1/tenants/current/by-slug/{slug}", tenant.GetBySlug(tenants))
	return r
}

func corsAllow(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Forwarded-Host, Host")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
