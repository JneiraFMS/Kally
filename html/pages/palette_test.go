package pages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// configRelPath / cssRelPath : chemins depuis le répertoire du package.
const (
	configRelPath = "../../tailwind.config.js"
	cssRelPath    = "../../web/static/tailwind.css"
)

// themeTokenNames : les tokens de palette couverts par les deux thèmes
// (map dark/light + :root + [data-theme="light"]).
var themeTokenNames = []string{
	"g-bg", "g-widget", "g-block", "g-hover", "g-border", "g-sep",
	"g-code", "g-accent", "g-accent-dim", "g-accent-text", "g-cyan",
	"g-hi", "g-text", "g-base", "g-muted", "g-subdue", "g-negative",
	"g-negative-text", "g-positive", "g-positive-text", "g-ink",
	"g-syn-keyword", "g-syn-string", "g-syn-comment", "g-syn-error",
}

// wantPaletteValues : valeurs canoniques attendues du thème dark (la base
// actuelle, palette Kally « console » : obsidienne + indigo) — dérive
// interdite. Équivalents Tailwind notés en commentaire pour le tuning.
var wantPaletteValues = map[string]string{
	"g-bg":            "#0b0f17", // bg-slate-950 observe
	"g-widget":        "#111726", // carte
	"g-block":         "#151d2e", // bloc surélevé
	"g-hover":         "#1d2739",
	"g-border":        "#1e293b", // border-slate-800
	"g-sep":           "#2a3648",
	"g-code":          "#0c1120",
	"g-accent":        "#4f46e5", // bg-indigo-600
	"g-accent-dim":    "#232347",
	"g-accent-text":   "#a5b4fc", // text-indigo-300
	"g-cyan":          "#818cf8", // focus ring (indigo-400)
	"g-hi":            "#f1f5f9", // text-slate-100
	"g-text":          "#94a3b8", // text-slate-400
	"g-base":          "#cbd5e1", // text-slate-300
	"g-muted":         "#64748b", // text-slate-500 (décoratif)
	"g-subdue":        "#475569", // text-slate-600 (décoratif)
	"g-negative":      "#ef4444", // red-500
	"g-negative-text": "#f87171", // text-red-400
	"g-positive":      "#10b981", // emerald-500
	"g-positive-text": "#6ee7b7", // text-emerald-300

	"g-ink": "#ffffff",

	"g-syn-keyword": "#a5b4fc",
	"g-syn-string":  "#6ee7b7",
	"g-syn-comment": "#a6b5ca",
	"g-syn-error":   "#fca5a5",
}

// wantLightValues : mode clair — slate-50 / blanc / indigo-600, accents
// assombris pour la lisibilité (écho du canon « Vercel docs » clair).
var wantLightValues = map[string]string{
	"g-bg":            "#f8fafc",
	"g-widget":        "#ffffff",
	"g-block":         "#f1f5f9",
	"g-hover":         "#e2e8f0",
	"g-border":        "#e2e8f0",
	"g-sep":           "#cbd5e1",
	"g-code":          "#f1f5f9",
	"g-accent":        "#4f46e5",
	"g-accent-dim":    "#e0e7ff",
	"g-accent-text":   "#4338ca",
	"g-cyan":          "#4f46e5",
	"g-hi":            "#0f172a",
	"g-text":          "#475569",
	"g-base":          "#334155",
	"g-muted":         "#94a3b8",
	"g-subdue":        "#cbd5e1",
	"g-negative":      "#dc2626",
	"g-negative-text": "#b91c1c",
	"g-positive":      "#059669",
	"g-positive-text": "#065f46",

	"g-ink": "#ffffff",

	"g-syn-keyword": "#4338ca",
	"g-syn-string":  "#047857",
	"g-syn-comment": "#475569",
	"g-syn-error":   "#b91c1c",
}

// TestThemePalettes : les deux palettes couvrent exactement les mêmes
// tokens, avec les valeurs canoniques attendues.
func TestThemePalettes(t *testing.T) {
	for _, name := range themeTokenNames {
		if paletteDark[name] != wantPaletteValues[name] {
			t.Errorf("paletteDark[%s] = %q, want %q", name, paletteDark[name], wantPaletteValues[name])
		}
		if paletteLight[name] != wantLightValues[name] {
			t.Errorf("paletteLight[%s] = %q, want %q", name, paletteLight[name], wantLightValues[name])
		}
	}
	for _, p := range []ThemeVars{paletteDark, paletteLight} {
		if len(p) != len(themeTokenNames) {
			t.Errorf("palette sans couverture complète : %d tokens, want %d", len(p), len(themeTokenNames))
		}
	}
}

// TestTailwindConfigUsesVars : tailwind.config.js doit référencer les
// variables CSS (le thème se bascule au runtime), plus le scan du Go.
func TestTailwindConfigUsesVars(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean(configRelPath))
	if err != nil {
		t.Fatalf("tailwind.config.js introuvable: %v", err)
	}
	config := string(data)

	for _, name := range themeTokenNames {
		want := "'" + name + "': 'var(--" + name + ")'"
		if !strings.Contains(config, want) {
			t.Errorf("tailwind.config.js sans %s", want)
		}
	}
	for _, pattern := range []string{"./main.go", "./html/**/*.go", "./web/**/*.go"} {
		if !strings.Contains(config, "'"+pattern+"'") {
			t.Errorf("tailwind.config.js : content sans %s", pattern)
		}
	}
}

// TestGeneratedStylesheetUsesVars : la feuille générée ne doit pas coder en
// dur l'hexapole couleur (sinon la bascule claire est ignorée).
func TestGeneratedStylesheetUsesVars(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean(cssRelPath))
	if err != nil {
		t.Fatalf("web/static/tailwind.css introuvable (lancer tools/build-css.sh): %v", err)
	}
	css := string(data)
	if strings.Contains(css, "0b0f17") {
		t.Errorf("feuille générée avec hex de fond en dur : le mode clair ne peut pas s'appliquer")
	}
	for _, rule := range []string{"--g-widget", "var(--g-hi)", ".font-mono", ".hidden"} {
		if !strings.Contains(css, rule) {
			t.Errorf("tailwind.css généré sans %s", rule)
		}
	}
	for _, rule := range []string{".w-10", ".shrink-0"} {
		if !strings.Contains(css, rule) {
			t.Errorf("tailwind.css généré sans la classe utilisée %s", rule)
		}
	}
}
