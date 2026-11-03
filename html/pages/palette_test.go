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
	"g-negative-text", "g-positive", "g-ink",
	"g-syn-keyword", "g-syn-string", "g-syn-comment", "g-syn-error",
}

// wantPaletteValues : valeurs Canonique attendues du thème dark (la base
// actuelle, Solarized dark) — dérive interdite.
var wantPaletteValues = map[string]string{
	"g-bg":            "#002b36",
	"g-widget":        "#073642",
	"g-block":         "#12373f",
	"g-hover":         "#214c56",
	"g-border":        "#1b444f",
	"g-sep":           "#28515a",
	"g-code":          "#24545e",
	"g-accent":        "#b58900",
	"g-accent-dim":    "#0c464f",
	"g-accent-text":   "#d7ad4a",
	"g-cyan":          "#2aa198",
	"g-hi":            "#eee8d5",
	"g-text":          "#93a1a1",
	"g-base":          "#839496",
	"g-muted":         "#657b83",
	"g-subdue":        "#586e75",
	"g-negative":      "#dc322f",
	"g-negative-text": "#f08b88",
	"g-positive":      "#859900",
	"g-ink":           "#002b36",

	"g-syn-keyword": "#c3d96b",
	"g-syn-string":  "#7ed3c9",
	"g-syn-comment": "#b9c6c6",
	"g-syn-error":   "#f6b0ad",
}

// wantLightValues : mode clair Solarized Light — échelle inversée canonique
// (fond base3, textes base01/base00) ; accents identiques au mode sombre ;
// les 4 tons intermédiaires (block/hover/border/sep) sont des dérivés
// blanchâtres, à valider par capture d'écran.
var wantLightValues = map[string]string{
	"g-bg":            "#fdf6e3",
	"g-widget":        "#eee8d5",
	"g-block":         "#e9e4d0",
	"g-hover":         "#ded9c2",
	"g-border":        "#d6d0b9",
	"g-sep":           "#c8c29f",
	"g-code":          "#fdf6e3",
	"g-accent":        "#b58900",
	"g-accent-dim":    "#f5ecc9",
	"g-accent-text":   "#7a5c00",
	"g-cyan":          "#2aa198",
	"g-hi":            "#073642",
	"g-text":          "#52686f",
	"g-base":          "#657b83",
	"g-muted":         "#839496",
	"g-subdue":        "#93a1a1",
	"g-negative":      "#dc322f",
	"g-negative-text": "#a0211e",
	"g-positive":      "#859900",
	"g-ink":           "#002b36",

	"g-syn-keyword": "#5c6a00",
	"g-syn-string":  "#0f6f68",
	"g-syn-comment": "#52686f",
	"g-syn-error":   "#a0211e",
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
	if strings.Contains(css, "002b36") {
		t.Errorf("feuille générée avec hex Solarized en dur : le mode clair ne peut pas s'appliquer")
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
