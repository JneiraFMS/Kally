package web

import (
	"embed"
	"log"
	"net/http"
)

// assetTailwindCSS est la feuille de styles Tailwind générée hors-ligne
// (tools/build-css.sh, palette de tailwind.config.js) : plus de CDN
// (offline, pas de FOUC, rendu complet même sans JS embarqué du CDN).
//
//go:embed static/tailwind.css
var assetTailwindCSS []byte

// assetFonts : les polices variables embarquées (Plus Jakarta Sans,
// JetBrains Mono) — typographie self-hosted, aucune dépendance réseau au
// chargement. Les noms sont figés au build : le contenu est caché
// immuablement.
//
//go:embed assets/fonts/*.woff2
var assetFonts embed.FS

// fontPathPrefix : chemin des fichiers dans l'embed.FS (routes = ce
// chemin sans le préfixe).
const fontPathPrefix = "assets/fonts/"

// ServeStylesheet sert la feuille /assets/tailwind.css.
func ServeStylesheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store") // la feuille est régénérée entre builds
	if _, err := w.Write(assetTailwindCSS); err != nil {
		log.Printf("serve Tailwind stylesheet: %v", err)
	}
}

// ServeFont sert un fichier de police embarqué (/assets/fonts/{name}).
// Le nom est contraint à l'ensemble embarqué : tout autre nom est 404, et
// aucune lecture du système de fichiers hôte n'a lieu. Cache immuable car
// le binaire embarque une copie finale.
func ServeFont(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	data, err := assetFonts.ReadFile(fontPathPrefix + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "font/woff2")
	// nosniff : le type de contenu déclaré est la seule interprétation
	// possible (police de sécuité passive).
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if _, err := w.Write(data); err != nil {
		log.Printf("serve font %s: %v", name, err)
	}
}
