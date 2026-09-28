package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/tenant"
)

type mapTenant struct {
	byHost map[string]*tenant.Public
}

func (m mapTenant) Resolve(_ context.Context, host, slug string) (*tenant.Public, error) {
	if t, ok := m.byHost[host]; ok {
		return t, nil
	}
	for _, t := range m.byHost {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, tenant.ErrNotFound
}

func twoTenantMux(t *testing.T) (http.Handler, *tenant.Public, *tenant.Public, string) {
	t.Helper()
	rowA, tnA, secret := testStaff(t)
	tnB := &tenant.Public{
		ID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Slug: "demo-b",
		Host: "demo-b.localhost",
		Name: "Demo B",
	}
	hash, err := HashPassword("changeme_staff")
	if err != nil {
		t.Fatal(err)
	}
	rowB := &StaffRow{
		ID:           uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		TenantID:     tnB.ID,
		Email:        "owner@demo-b.local",
		Name:         "Dueño B",
		Role:         "owner",
		PasswordHash: hash,
		Active:       true,
	}
	dir := newMemDir(rowA)
	dir.staff[rowB.ID] = rowB
	dir.byEmail[rowB.Email] = rowB
	res := mapTenant{byHost: map[string]*tenant.Public{tnA.Host: tnA, tnB.Host: tnB}}
	svc := NewService(dir, res, secret)
	r := chi.NewRouter()
	r.Use(platmw.IgnoreClientTenantID)
	r.Post("/v1/auth/login", Login(svc))
	r.Group(func(gr chi.Router) {
		gr.Use(platmw.RequireStaff(secret, LookupTenantID(res)))
		gr.Get("/v1/auth/me", Me(svc))
		gr.Get("/v1/staff", ListStaff(svc))
	})
	return r, tnA, tnB, secret
}

func TestJWTofAOnHostBIs401(t *testing.T) {
	mux, tnA, tnB, _ := twoTenantMux(t)
	token := bearer(t, mux, tnA.Host, "owner@demo-a.local", "changeme_staff")

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Host = tnB.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("A token on B host: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Host = tnA.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("A token on A host: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAuthRoutesIgnoreXTenantID(t *testing.T) {
	mux, tnA, tnB, _ := twoTenantMux(t)
	token := bearer(t, mux, tnA.Host, "owner@demo-a.local", "changeme_staff")

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Host = tnA.Host
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-Id", tnB.ID.String())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me %d %s", rec.Code, rec.Body.String())
	}
	var me StaffPublic
	if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.TenantSlug != tnA.Slug {
		t.Fatalf("X-Tenant-Id must not switch tenant: %+v", me)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/staff", nil)
	req.Host = tnA.Host
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-Id", tnB.ID.String())
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("staff %d %s", rec.Code, rec.Body.String())
	}
	var list []StaffPublic
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	for _, s := range list {
		if s.Email == "owner@demo-b.local" {
			t.Fatal("listed B staff via X-Tenant-Id")
		}
	}

	body, _ := json.Marshal(map[string]string{"email": "owner@demo-a.local", "password": "changeme_staff"})
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Host = tnA.Host
	req.Header.Set("X-Tenant-Id", tnB.ID.String())
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login with X-Tenant-Id of B: %d %s", rec.Code, rec.Body.String())
	}
}
