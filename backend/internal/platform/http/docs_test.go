package httphandler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
)

func TestSwaggerDocsServeOpenAPI(t *testing.T) {
	r := NewRouter(&database.DB{}, nil, "")
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("openapi:")) {
		t.Fatal("expected openapi document")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("/v1/auth/login")) {
		t.Fatal("expected login path")
	}

	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("docs %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("swagger-ui")) {
		t.Fatal("expected swagger ui")
	}
}
