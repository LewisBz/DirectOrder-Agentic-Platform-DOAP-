package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func bearer(t *testing.T, mux http.Handler, host, email, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Host = host
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %d %s", rec.Code, rec.Body.String())
	}
	var pair TokenPair
	if err := json.Unmarshal(rec.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	return pair.AccessToken
}

func TestStaffCRUDContract(t *testing.T) {
	row, tn, secret := testStaff(t)
	mux := testMux(row, tn, secret)
	token := bearer(t, mux, tn.Host, row.Email, "changeme_staff")

	create, _ := json.Marshal(map[string]string{
		"email":    "caja@demo-a.local",
		"password": "cajero123",
		"name":     "Caja Uno",
		"role":     "cashier",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/staff", bytes.NewReader(create))
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var created StaffPublic
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Role != "cashier" || created.Email != "caja@demo-a.local" {
		t.Fatalf("%+v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/staff", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d", rec.Code)
	}

	ownerCreate, _ := json.Marshal(map[string]string{
		"email": "otro@demo-a.local", "password": "cajero123", "name": "X", "role": "owner",
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/staff", bytes.NewReader(ownerCreate))
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("owner role %d %s", rec.Code, rec.Body.String())
	}

	dup, _ := json.Marshal(map[string]string{
		"email": "caja@demo-a.local", "password": "cajero123", "name": "Dup", "role": "kitchen",
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/staff", bytes.NewReader(dup))
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("dup %d", rec.Code)
	}

	patch, _ := json.Marshal(map[string]string{"name": "Caja Editada"})
	req = httptest.NewRequest(http.MethodPatch, "/v1/staff/"+created.ID.String(), bytes.NewReader(patch))
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/staff/"+created.ID.String()+"/deactivate", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("deact %d", rec.Code)
	}

	cajaTokBody, _ := json.Marshal(map[string]string{"email": "caja@demo-a.local", "password": "cajero123"})
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(cajaTokBody))
	req.Host = tn.Host
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("inactive login %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/staff/"+created.ID.String()+"/reactivate", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("react %d", rec.Code)
	}

	pw, _ := json.Marshal(map[string]string{"password": "nuevaclave"})
	req = httptest.NewRequest(http.MethodPatch, "/v1/staff/"+created.ID.String(), bytes.NewReader(pw))
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch pw %d %s", rec.Code, rec.Body.String())
	}

	cajaTok := bearer(t, mux, tn.Host, "caja@demo-a.local", "nuevaclave")
	req = httptest.NewRequest(http.MethodGet, "/v1/staff", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+cajaTok)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cashier list %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/staff/"+row.ID.String()+"/deactivate", nil)
	req.Host = tn.Host
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("deact owner %d %s", rec.Code, rec.Body.String())
	}
}
