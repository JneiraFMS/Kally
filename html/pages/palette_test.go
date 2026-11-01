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

// TestPaletteSolarizedValues : les constantes de la palette correspondent
// aux valeurs Solarized attendues (dérive impossible sans test rouge).
func TestPaletteSolarizedValues(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"base03", solarizedBase03, "#002b36"},
		{"base02", solarizedBase02, "#073642"},
		{"base01", solarizedBase01, "#586e75"},
		{"base00", solarizedBase00, "#657b83"},
		{"base0", solarizedBase0, "#839496"},
		{"base1", solarizedBase1, "#93a1a1"},
		{"base2", solarizedBase2, "#eee8d5"},
		{"yellow", solarizedYellow, "#b58900"},
		{"cyan", solarizedCyan, "#2aa198"},
		{"red", solarizedRed, "#dc322f"},
		{"green", solarizedGreen, "#859900"},
		{"sep", solarizedSep, "#28515a"},
		{"block", solarizedBlock, "#15414d"},
		{"border", solarizedBorder, "#1b444f"},
		{"hover", solarizedHover, "#214c56"},
		{"accentDim", solarizedAccentDim, "#0c464f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestTailwindConfigMatchesPalette : tailwind.config.js doit refléter les
// constantes Go (palette.go) — sync des deux sources.
func TestTailwindConfigMatchesPalette(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean(configRelPath))
	if err != nil {
		t.Fatalf("tailwind.config.js introuvable: %v", err)
	}
	config := string(data)

	pairs := map[string]string{
		solarizedBase02:    "g-widget",
		solarizedBlock:     "g-block",
		solarizedHover:     "g-hover",
		solarizedBorder:    "g-border",
		solarizedSep:       "g-sep",
		solarizedYellow:    "g-accent",
		solarizedAccentDim: "g-accent-dim",
		solarizedBase2:     "g-hi",
		solarizedBase1:     "g-text",
		solarizedBase0:     "g-base",
		solarizedBase00:    "g-muted",
		solarizedBase01:    "g-subdue",
		solarizedRed:       "g-negative",
		solarizedGreen:     "g-positive",
		// g-ink = texte inversé sur l'accent jaune (= fond)
		solarizedBase03: "g-ink",
	}
	for hex, token := range pairs {
		if !strings.Contains(config, "'"+token+"': '"+hex+"'") {
			t.Errorf("tailwind.config.js : %s (%s) manquant", token, hex)
		}
	}
	// g-bg partage la même valeur que g-ink (base03)
	if !strings.Contains(config, "'g-bg': '"+solarizedBase03+"'") {
		t.Errorf("tailwind.config.js : g-bg manquant")
	}

	// Le scan doit couvrir les sources contenant les classes utilitaires.
	for _, pattern := range []string{"./main.go", "./html/**/*.go", "./web/**/*.go"} {
		if !strings.Contains(config, "'"+pattern+"'") {
			t.Errorf("tailwind.config.js : content sans %s", pattern)
		}
	}
}

// TestGeneratedStylesheetComplete : la feuille générée doit exister et
// contenir les utilitaires clés (palette, classes arbitraires remplacées).
func TestGeneratedStylesheetComplete(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean(cssRelPath))
	if err != nil {
		t.Fatalf("web/static/tailwind.css introuvable (lancer tools/build-css.sh): %v", err)
	}
	css := string(data)
	for _, cls := range []string{
		".text-g-hi", ".bg-g-widget", ".text-g-ink", ".bg-g-accent",
		".placeholder-g-muted", ".font-mono", ".hidden",
	} {
		if !strings.Contains(css, cls) {
			t.Errorf("tailwind.css généré sans la règle %s", cls)
		}
	}
	// Preflight présent (bases Tailwind)
	if !strings.Contains(css, "-webkit-font-smoothing") {
		t.Errorf("preflight Tailwind manquant dans la feuille générée")
	}
}
