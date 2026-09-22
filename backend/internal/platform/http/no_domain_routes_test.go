package httphandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
)

func TestNoOrderCartPaymentAgentRoutes(t *testing.T) {
	r := NewRouter(&database.DB{})
	for _, path := range []string{"/v1/orders", "/v1/cart", "/v1/payments", "/v1/agent"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: got %d want 404", path, rec.Code)
		}
	}
}
