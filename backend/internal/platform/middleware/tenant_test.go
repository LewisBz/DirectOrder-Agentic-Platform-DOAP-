package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIgnoreClientTenantIDStripsHeaderOnAuthPaths(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Tenant-Id"); got != "" {
			t.Errorf("X-Tenant-Id still present: %s", got)
		}
		if HostFromContext(r.Context()) != "demo-a.localhost" {
			t.Errorf("host %q", HostFromContext(r.Context()))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Host = "demo-a.localhost"
	req.Header.Set("X-Tenant-Id", "22222222-2222-2222-2222-222222222222")
	rec := httptest.NewRecorder()
	IgnoreClientTenantID(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
}
