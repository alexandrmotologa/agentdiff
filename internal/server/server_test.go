package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerHealthAndStatic(t *testing.T) {
	srv := NewServer("", "")
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	mux.Handle("/", srv.FileServerHandler())

	// Test /api/health
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /api/health, got %d", rec.Code)
	}

	// Test embedded index.html fallback
	reqSpa := httptest.NewRequest("GET", "/dashboard", nil)
	recSpa := httptest.NewRecorder()
	mux.ServeHTTP(recSpa, reqSpa)

	if recSpa.Code != http.StatusOK {
		t.Fatalf("expected status 200 from SPA root, got %d", recSpa.Code)
	}
}

func TestGenerateStandaloneHTML(t *testing.T) {
	tmpDir := t.TempDir()
	html, err := GenerateStandaloneHTML(tmpDir)
	if err != nil {
		t.Fatalf("GenerateStandaloneHTML failed: %v", err)
	}
	if len(html) == 0 {
		t.Fatalf("expected non-empty html string")
	}
}
