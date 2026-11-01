package web

import (
	_ "embed"
	"net/http"
)

// assetTailwindCSS est la feuille de styles Tailwind générée hors-ligne
// (tools/build-css.sh, palette Solarized de tailwind.config.js) : plus de
// CDN (offline, pas de FOUC, rendu complet même sans JS embarqué du CDN).
//
//go:embed static/tailwind.css
var assetTailwindCSS []byte

// ServeStylesheet sert la feuille /assets/tailwind.css.
func ServeStylesheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store") // la feuille est régénérée entre builds
	_, _ = w.Write(assetTailwindCSS)
}
