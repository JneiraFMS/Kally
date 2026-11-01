package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestServeStylesheet : la feuille Tailwind générée est servie en CSS avec
// les utilitaires attendus (pas de CDN en prod).
func TestServeStylesheet(t *testing.T) {
	req := httptest.NewRequest("GET", "/assets/tailwind.css", nil)
	w := httptest.NewRecorder()

	ServeStylesheet(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ServeStylesheet() status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("Content-Type = %q, want text/css", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, ".text-g-ink") {
		t.Errorf("feuille générée sans .text-g-ink (recompiler via tools/build-css.sh)")
	}
	if len(body) < 1000 {
		t.Errorf("feuille générée trop courte : %d octets", len(body))
	}
}
