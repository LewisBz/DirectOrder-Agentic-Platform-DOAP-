package tenant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
)

type fakeResolver struct {
	byHost map[string]*Public
	bySlug map[string]*Public
}

func (f fakeResolver) Resolve(_ context.Context, host, slug string) (*Public, error) {
	if host != "" {
		if t, ok := f.byHost[host]; ok {
			return t, nil
		}
	}
	if slug != "" {
		if t, ok := f.bySlug[slug]; ok {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

func demoResolver() fakeResolver {
	a := &Public{
		ID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Slug: "demo-a", Host: "demo-a.localhost", Name: "Demo A",
		Currency: "COP", TaxName: "IVA", Timezone: "America/Bogota",
	}
	b := &Public{
		ID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Slug: "demo-b", Host: "demo-b.localhost", Name: "Demo B",
		Currency: "COP", TaxName: "IVA", Timezone: "America/Bogota",
	}
	return fakeResolver{
		byHost: map[string]*Public{"demo-a.localhost": a, "demo-b.localhost": b},
		bySlug: map[string]*Public{"demo-a": a, "demo-b": b},
	}
}

func testRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(platmw.IgnoreClientTenantID)
	res := demoResolver()
	r.Get("/v1/tenants/current", GetCurrent(res))
	r.Get("/v1/tenants/current/by-slug/{slug}", GetBySlug(res))
	return r
}

func TestGetCurrentByHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/current", nil)
	req.Host = "demo-a.localhost"
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body Public
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Slug != "demo-a" || body.Currency != "COP" || body.TaxName != "IVA" || body.Timezone != "America/Bogota" {
		t.Fatalf("%+v", body)
	}
}

func TestGetCurrentXForwardedHostWins(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/current", nil)
	req.Host = "demo-a.localhost"
	req.Header.Set("X-Forwarded-Host", "demo-b.localhost")
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, req)
	var body Public
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body.Slug != "demo-b" {
		t.Fatalf("status %d slug %s", rec.Code, body.Slug)
	}
}

func TestGetCurrentIgnoresXTenantID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/current", nil)
	req.Host = "demo-a.localhost"
	req.Header.Set("X-Tenant-Id", "22222222-2222-2222-2222-222222222222")
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, req)
	var body Public
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body.Slug != "demo-a" {
		t.Fatalf("status %d slug %s", rec.Code, body.Slug)
	}
}

func TestGetCurrentUnknownHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/current", nil)
	req.Host = "unknown.localhost"
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	var errBody map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &errBody)
	if errBody["error"] != "tenant_not_found" {
		t.Fatalf("%v", errBody)
	}
}

func TestGetBySlug(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/current/by-slug/demo-a", nil)
	req.Host = "localhost"
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, req)
	var body Public
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body.Slug != "demo-a" {
		t.Fatalf("status %d slug %s", rec.Code, body.Slug)
	}
}

func TestHostWinsOverConflictingSlug(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/current/by-slug/demo-b", nil)
	req.Host = "demo-a.localhost"
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, req)
	var body Public
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body.Slug != "demo-a" {
		t.Fatalf("host should win: status %d slug %s", rec.Code, body.Slug)
	}
}
