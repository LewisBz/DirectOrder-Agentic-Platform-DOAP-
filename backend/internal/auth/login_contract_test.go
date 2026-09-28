package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/tenant"
)

type memDir struct {
	mu      sync.Mutex
	staff   map[uuid.UUID]*StaffRow
	byEmail map[string]*StaffRow
	refresh map[string]memRefresh
	guests  map[string]uuid.UUID
}

type memRefresh struct {
	staffID uuid.UUID
	exp     time.Time
	revoked bool
}

func newMemDir(row *StaffRow) *memDir {
	return &memDir{
		staff:   map[uuid.UUID]*StaffRow{row.ID: row},
		byEmail: map[string]*StaffRow{row.Email: row},
		refresh: map[string]memRefresh{},
		guests:  map[string]uuid.UUID{},
	}
}

func (m *memDir) FindByEmail(_ context.Context, tenantID uuid.UUID, email string) (*StaffRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.byEmail[strings.ToLower(email)]
	if row == nil || row.TenantID != tenantID {
		return nil, ErrInvalidCredentials
	}
	cp := *row
	return &cp, nil
}

func (m *memDir) GetByID(_ context.Context, tenantID, staffID uuid.UUID) (*StaffRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.staff[staffID]
	if row == nil || row.TenantID != tenantID {
		return nil, ErrInvalidCredentials
	}
	cp := *row
	return &cp, nil
}

func (m *memDir) InsertRefresh(_ context.Context, _, staffID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refresh[tokenHash] = memRefresh{staffID: staffID, exp: expiresAt}
	return nil
}

func (m *memDir) LookupRefresh(_ context.Context, _ uuid.UUID, tokenHash string) (uuid.UUID, time.Time, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.refresh[tokenHash]
	if !ok {
		return uuid.Nil, time.Time{}, false, ErrInvalidCredentials
	}
	return r.staffID, r.exp, r.revoked, nil
}

func (m *memDir) RevokeRefresh(_ context.Context, _ uuid.UUID, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.refresh[tokenHash]
	r.revoked = true
	m.refresh[tokenHash] = r
	return nil
}

func (m *memDir) RevokeAllRefresh(_ context.Context, _, staffID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, r := range m.refresh {
		if r.staffID == staffID {
			r.revoked = true
			m.refresh[k] = r
		}
	}
	return nil
}

func (m *memDir) List(_ context.Context, tenantID uuid.UUID) ([]StaffRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []StaffRow
	for _, row := range m.staff {
		if row.TenantID == tenantID {
			cp := *row
			out = append(out, cp)
		}
	}
	return out, nil
}

func (m *memDir) Create(_ context.Context, tenantID uuid.UUID, email, name, role, passwordHash string) (*StaffRow, error) {
	if role != "cashier" && role != "kitchen" {
		return nil, ErrOwnerRole
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	email = strings.ToLower(strings.TrimSpace(email))
	if _, ok := m.byEmail[email]; ok {
		return nil, ErrConflict
	}
	row := &StaffRow{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		Name:         name,
		Role:         role,
		PasswordHash: passwordHash,
		Active:       true,
	}
	m.staff[row.ID] = row
	m.byEmail[email] = row
	cp := *row
	return &cp, nil
}

func (m *memDir) Update(_ context.Context, tenantID, staffID uuid.UUID, email, name, role string, passwordHash *string) (*StaffRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.staff[staffID]
	if row == nil || row.TenantID != tenantID {
		return nil, ErrStaffNotFound
	}
	if row.Role == "owner" {
		return nil, ErrOwnerLocked
	}
	if role != "cashier" && role != "kitchen" {
		return nil, ErrOwnerRole
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if other := m.byEmail[email]; other != nil && other.ID != staffID {
		return nil, ErrConflict
	}
	delete(m.byEmail, row.Email)
	row.Email = email
	row.Name = name
	row.Role = role
	if passwordHash != nil {
		row.PasswordHash = *passwordHash
		for k, r := range m.refresh {
			if r.staffID == staffID {
				r.revoked = true
				m.refresh[k] = r
			}
		}
	}
	m.byEmail[email] = row
	cp := *row
	return &cp, nil
}

func (m *memDir) SetActive(_ context.Context, tenantID, staffID uuid.UUID, active bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.staff[staffID]
	if row == nil || row.TenantID != tenantID {
		return ErrStaffNotFound
	}
	if row.Role == "owner" {
		return ErrOwnerLocked
	}
	row.Active = active
	if !active {
		for k, r := range m.refresh {
			if r.staffID == staffID {
				r.revoked = true
				m.refresh[k] = r
			}
		}
	}
	return nil
}

func (m *memDir) FindGuest(_ context.Context, tenantID uuid.UUID, tokenHash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	got, ok := m.guests[tokenHash]
	return ok && got == tenantID, nil
}

func (m *memDir) InsertGuest(_ context.Context, tenantID uuid.UUID, tokenHash, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.guests[tokenHash] = tenantID
	return nil
}

func (m *memDir) TouchGuest(_ context.Context, tenantID uuid.UUID, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.guests[tokenHash] != tenantID {
		return ErrStaffNotFound
	}
	return nil
}

type staticTenant struct{ t *tenant.Public }

func (s staticTenant) Resolve(_ context.Context, host, slug string) (*tenant.Public, error) {
	if host == s.t.Host || slug == s.t.Slug {
		return s.t, nil
	}
	return nil, tenant.ErrNotFound
}

func testStaff(t *testing.T) (*StaffRow, *tenant.Public, string) {
	t.Helper()
	hash, err := HashPassword("changeme_staff")
	if err != nil {
		t.Fatal(err)
	}
	tn := &tenant.Public{
		ID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Slug: "demo-a",
		Host: "demo-a.localhost",
		Name: "Demo A",
	}
	row := &StaffRow{
		ID:           uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		TenantID:     tn.ID,
		Email:        "owner@demo-a.local",
		Name:         "Dueño A",
		Role:         "owner",
		PasswordHash: hash,
		Active:       true,
	}
	return row, tn, "changeme_jwt_secret_do_not_use_prod_xx"
}

func testMux(row *StaffRow, tn *tenant.Public, secret string) http.Handler {
	svc := NewService(newMemDir(row), staticTenant{t: tn}, secret)
	r := chi.NewRouter()
	r.Use(platmw.IgnoreClientTenantID)
	r.Post("/v1/auth/login", Login(svc))
	r.Post("/v1/auth/refresh", Refresh(svc))
	r.Group(func(gr chi.Router) {
		gr.Use(platmw.RequireStaff(secret, LookupTenantID(staticTenant{t: tn})))
		gr.Post("/v1/auth/logout", Logout(svc))
		gr.Get("/v1/auth/me", Me(svc))
		gr.Get("/v1/staff", ListStaff(svc))
		gr.Post("/v1/staff", CreateStaff(svc))
		gr.Patch("/v1/staff/{id}", PatchStaff(svc))
		gr.Post("/v1/staff/{id}/deactivate", DeactivateStaff(svc))
		gr.Post("/v1/staff/{id}/reactivate", ReactivateStaff(svc))
	})
	return r
}

func TestLoginMeRefreshLogoutContract(t *testing.T) {
	row, tn, secret := testStaff(t)
	mux := testMux(row, tn, secret)

	loginBody, _ := json.Marshal(map[string]string{"email": row.Email, "password": "changeme_staff"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	req.Host = tn.Host
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status %d body %s", rec.Code, rec.Body.String())
	}
	var pair TokenPair
	if err := json.Unmarshal(rec.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || pair.TokenType != "Bearer" || pair.ExpiresIn != 900 {
		t.Fatalf("pair %+v", pair)
	}

	bad, _ := json.Marshal(map[string]string{"email": row.Email, "password": "wrong-pass"})
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(bad))
	req.Host = tn.Host
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad login %d", rec.Code)
	}
	var errBody map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &errBody)
	if errBody["error"] != "invalid_credentials" {
		t.Fatalf("error %v", errBody)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me %d %s", rec.Code, rec.Body.String())
	}
	var me StaffPublic
	if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.Role != "owner" || me.TenantSlug != "demo-a" || me.Name != "Dueño A" {
		t.Fatalf("me %+v", me)
	}

	refBody, _ := json.Marshal(map[string]string{"refresh_token": pair.RefreshToken})
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(refBody))
	req.Host = tn.Host
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh %d %s", rec.Code, rec.Body.String())
	}
	var pair2 TokenPair
	_ = json.Unmarshal(rec.Body.Bytes(), &pair2)
	if pair2.RefreshToken == pair.RefreshToken {
		t.Fatal("refresh must rotate")
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(refBody))
	req.Host = tn.Host
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("old refresh should fail %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+pair2.AccessToken)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout %d", rec.Code)
	}
}

func TestPasswordCost12(t *testing.T) {
	h, err := HashPassword("changeme_staff")
	if err != nil {
		t.Fatal(err)
	}
	c, err := bcrypt.Cost([]byte(h))
	if err != nil {
		t.Fatal(err)
	}
	if c != PasswordCost {
		t.Fatalf("cost %d", c)
	}
}
