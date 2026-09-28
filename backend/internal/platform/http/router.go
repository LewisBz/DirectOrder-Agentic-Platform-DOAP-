package httphandler

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/auth"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/tenant"
)

func NewRouter(db *database.DB, tenants tenant.Resolver, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(corsAllow)
	r.Use(platmw.IgnoreClientTenantID)
	r.Get("/healthz", Healthz())
	r.Get("/readyz", Readyz(db))
	r.Get("/openapi.yaml", OpenAPISpec())
	r.Get("/docs", SwaggerUI())
	r.Get("/docs/", SwaggerUI())
	r.Get("/swagger", SwaggerUI())
	r.Get("/v1/tenants/current", tenant.GetCurrent(tenants))
	r.Get("/v1/tenants/current/by-slug/{slug}", tenant.GetBySlug(tenants))

	authSvc := auth.NewService(auth.NewStore(db), tenants, jwtSecret)
	r.Post("/v1/auth/login", auth.Login(authSvc))
	r.Post("/v1/auth/refresh", auth.Refresh(authSvc))
	r.Post("/v1/guest/sessions", auth.EnsureGuest(authSvc))
	r.Group(func(gr chi.Router) {
		gr.Use(platmw.RequireStaff(jwtSecret, auth.LookupTenantID(tenants)))
		gr.Post("/v1/auth/logout", auth.Logout(authSvc))
		gr.Get("/v1/auth/me", auth.Me(authSvc))
		gr.Get("/v1/staff", auth.ListStaff(authSvc))
		gr.Post("/v1/staff", auth.CreateStaff(authSvc))
		gr.Patch("/v1/staff/{id}", auth.PatchStaff(authSvc))
		gr.Post("/v1/staff/{id}/deactivate", auth.DeactivateStaff(authSvc))
		gr.Post("/v1/staff/{id}/reactivate", auth.ReactivateStaff(authSvc))
	})
	return r
}

func corsAllow(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Forwarded-Host, Host, X-Guest-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
