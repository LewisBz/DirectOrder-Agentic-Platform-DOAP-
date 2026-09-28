package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/tenant"
)

func guestMux(t *testing.T) (http.Handler, *tenant.Public, *tenant.Public) {
	t.Helper()
	rowA, tnA, secret := testStaff(t)
	tnB := &tenant.Public{
		ID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Slug: "demo-b",
		Host: "demo-b.localhost",
		Name: "Demo B",
	}
	dir := newMemDir(rowA)
	res := mapTenant{byHost: map[string]*tenant.Public{tnA.Host: tnA, tnB.Host: tnB}}
	svc := NewService(dir, res, secret)
	r := chi.NewRouter()
	r.Use(platmw.IgnoreClientTenantID)
	r.Post("/v1/guest/sessions", EnsureGuest(svc))
	r.Group(func(gr chi.Router) {
		gr.Use(platmw.RequireStaff(secret, LookupTenantID(res)))
		gr.Get("/v1/staff", ListStaff(svc))
		gr.Get("/v1/auth/me", Me(svc))
	})
	return r, tnA, tnB
}

func parseGuest(t *testing.T, rec *httptest.ResponseRecorder) GuestSessionPublic {
	t.Helper()
	var gs GuestSessionPublic
	if err := json.Unmarshal(rec.Body.Bytes(), &gs); err != nil {
		t.Fatal(err)
	}
	return gs
}

func TestGuestEnsureContract(t *testing.T) {
	mux, tnA, tnB := guestMux(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/guest/sessions", bytes.NewReader([]byte(`{"channel":"pwa"}`)))
	req.Host = tnA.Host
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("new guest %d %s", rec.Code, rec.Body.String())
	}
	first := parseGuest(t, rec)
	if first.Token == "" || first.TenantSlug != tnA.Slug || first.Channel != "pwa" {
		t.Fatalf("%+v", first)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/guest/sessions", bytes.NewReader([]byte(`{"channel":"pwa"}`)))
	req.Host = tnA.Host
	req.Header.Set("X-Guest-Token", first.Token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("reuse %d %s", rec.Code, rec.Body.String())
	}
	reused := parseGuest(t, rec)
	if reused.Token != first.Token {
		t.Fatalf("expected same token, got %s", reused.Token)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/guest/sessions", bytes.NewReader([]byte(`{"channel":"pwa","token":"`+first.Token+`"}`)))
	req.Host = tnB.Host
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("other tenant %d %s", rec.Code, rec.Body.String())
	}
	other := parseGuest(t, rec)
	if other.Token == first.Token || other.TenantSlug != tnB.Slug {
		t.Fatalf("expected new token for B: %+v", other)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/guest/sessions", nil)
	req.Host = "unknown.localhost"
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown host %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/staff", nil)
	req.Host = tnA.Host
	req.Header.Set("X-Guest-Token", first.Token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("guest staff %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Host = tnA.Host
	req.Header.Set("Authorization", "Bearer "+first.Token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("guest as bearer %d", rec.Code)
	}
}
