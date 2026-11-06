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
// actuelle : fond type Dracula, cartes #21222c, accent pourpre #bd93f9) —
// dérive interdite. g-code est une couche PLUS PROFONDE que la carte ;
// g-negative-text / g-ink / g-syn-comment sont les variantes calibrées
// AA (>= 4.5:1) de leurs tons bruts.
var wantPaletteValues = map[string]string{
	"g-bg":            "#282a36", // Background
	"g-widget":        "#21222c", // carte
	"g-block":         "#343746", // surface secondaire
	"g-hover":         "#44475a", // Current Line / Hover
	"g-border":        "#44475a", // filet
	"g-sep":           "#6272a4", // separator
	"g-code":          "#17181f", // code : couche profonde
	"g-accent":        "#bd93f9", // Purple
	"g-accent-dim":    "#443653", // Purple muted
	"g-accent-text":   "#bd93f9", // Purple text
	"g-cyan":          "#8be9fd", // focus ring
	"g-hi":            "#f8f8f2", // Foreground
	"g-text":          "#e0e0e0", // Main text
	"g-base":          "#f8f8f2", // Base text
	"g-muted":         "#a6a6b3", // Muted text (décoratif)
	"g-subdue":        "#6272a4", // Subdued text (décoratif)
	"g-negative":      "#ff5555", // Red
	"g-negative-text": "#ff7d7d", // Light red AA sur g-block
	"g-positive":      "#50fa7b", // Green
	"g-positive-text": "#69ff94", // Light green

	"g-ink": "#282a36", // encre sombre sur l'accent pourpre (AA)

	"g-syn-keyword": "#ff79c6",
	"g-syn-string":  "#f1fa8c",
	"g-syn-comment": "#7c87b8",
	"g-syn-error":   "#ff5555",
}

// wantLightValues : mode clair — fond quasi-blanc, cartes blanches,
// accents assombris pour la lisibilité (>= 4.5:1 sur les surfaces claires).
var wantLightValues = map[string]string{
	"g-bg":            "#fcfcfd",
	"g-widget":        "#ffffff",
	"g-block":         "#f1f0f7",
	"g-hover":         "#e8e6f0",
	"g-border":        "#e0deeb",
	"g-sep":           "#c8c5d8",
	"g-code":          "#f1f0f7",
	"g-accent":        "#8657c9", // pourpre assombri : encre blanche AA
	"g-accent-dim":    "#eee5ff",
	"g-accent-text":   "#7044b8",
	"g-cyan":          "#008fb2",
	"g-hi":            "#282a36",
	"g-text":          "#44475a",
	"g-base":          "#343746",
	"g-muted":         "#77778b",
	"g-subdue":        "#a6a6b3",
	"g-negative":      "#d83b52",
	"g-negative-text": "#b52b43",
	"g-positive":      "#168b55",
	"g-positive-text": "#087443",

	"g-ink": "#ffffff",

	"g-syn-keyword": "#a626a4",
	"g-syn-string":  "#136f34",
	"g-syn-comment": "#5f5f74",
	"g-syn-error":   "#b02b41",
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
