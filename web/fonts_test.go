package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newFontMux reproduit le câblage réel de main.go pour les polices :
// une route GET /assets/fonts/{name} branchée sur ServeFont.
func newFontMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/fonts/{name}", ServeFont)
	return mux
}

// fontNames : les deux polices variables embarquées (axe de poids complet,
// découpé en sous-ensembles par le CDN Google Fonts à la génération).
var fontNames = []string{
	"plus-jakarta-sans.woff2",
	"jetbrains-mono.woff2",
}

// TestServeFont : chaque woff2 embarqué est servi avec le type de contenu
// correct, la signature binaire wOF2 et un cache immuable (nom figé au
// moment du build : le contenu ne peut pas changer au runtime).
func TestServeFont(t *testing.T) {
	mux := newFontMux()

	for _, name := range fontNames {
		req := httptest.NewRequest("GET", "/assets/fonts/"+name, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("%s : status = %d, want 200", name, w.Code)
			continue
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "font/woff2") {
			t.Errorf("%s : Content-Type = %q, want font/woff2", name, ct)
		}
		if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
			t.Errorf("%s : Cache-Control = %q, want immutable (le nom est figé au build)", name, cc)
		}
		if !bytes.HasPrefix(w.Body.Bytes(), []byte("wOF2")) {
			t.Errorf("%s : signature woff2 (wOF2) absente du corps servi", name)
		}
	}
}

// TestServeFontNotFound : un nom inconnu est 404 ; la traversal n'est
// jamais servie — le mux canonicalise le chemin (3xx) AVANT le handler, et
// aucun contenu du système de fichiers hôte ne doit fuiter dans la réponse.
func TestServeFontNotFound(t *testing.T) {
	mux := newFontMux()

	for _, name := range []string{"roboto.woff2", "../main.go", "sub/dir/x.woff2"} {
		req := httptest.NewRequest("GET", "/assets/fonts/"+name, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if code := w.Code; code == http.StatusOK {
			t.Errorf("font %q : le handler a répondu 200", name)
			continue
		}
		if name != "../main.go" && w.Code != http.StatusNotFound {
			t.Errorf("font %q : status = %d, want 404", name, w.Code)
			continue
		}
		if name == "../main.go" && strings.Contains(w.Body.String(), "package main") {
			t.Errorf("font %q : contenu du système de fichiers hôte servi", name)
		}
	}
}
