package pages

import (
	"fmt"
	"math"
	"testing"
)

// relativeLuminance calcule la luminance relative WCAG d'une couleur #rrggbb.
func relativeLuminance(hex string) float64 {
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		panic("couleur invalide: " + hex)
	}
	channel := func(v int) float64 {
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}

// contrastRatio retourne le ratio de contraste WCAG entre deux couleurs
// opaques #rrggbb (1:1 à 21:1).
func contrastRatio(fg, bg string) float64 {
	a := relativeLuminance(fg)
	b := relativeLuminance(bg)
	if a < b {
		a, b = b, a
	}
	return (a + 0.05) / (b + 0.05)
}

// TestPaletteContrast : chaque paire texte/fond réellement employée par
// l'interface atteint 4.5:1 (WCAG AA pour du texte de moins de 18,66 px).
// Les paires correspondent aux classes utilisées dans html/ et web/ :
//   - g-text  : corps, extraits, libellés, tags, actions secondaires ;
//   - g-hi    : titres, valeurs saisies ;
//   - g-accent-text : identifiants monospace et liens accentués ;
//   - g-negative-text : action destructive et messages d'erreur ;
//   - g-ink   : texte posé sur l'accent (boutons primaires).
func TestPaletteContrast(t *testing.T) {
	pairs := []struct {
		name   string
		fg, bg string
	}{
		{"corps sur fond", "g-text", "g-bg"},
		{"corps sur carte", "g-text", "g-widget"},
		{"corps sur bloc", "g-text", "g-block"},
		{"titre sur bloc", "g-hi", "g-block"},
		{"accent sur bloc", "g-accent-text", "g-block"},
		{"accent sur carte", "g-accent-text", "g-widget"},
		{"accent sur fond", "g-accent-text", "g-bg"},
		{"danger sur bloc", "g-negative-text", "g-block"},
		{"danger sur carte", "g-negative-text", "g-widget"},
		{"danger sur fond", "g-negative-text", "g-bg"},
		{"validation sur bloc", "g-positive-text", "g-block"},
		{"validation sur carte", "g-positive-text", "g-widget"},
		{"validation sur fond", "g-positive-text", "g-bg"},
		{"encre sur accent", "g-ink", "g-accent"},
		{"texte sur code", "g-hi", "g-code"},
		{"mot-clé sur code", "g-syn-keyword", "g-code"},
		{"chaîne sur code", "g-syn-string", "g-code"},
		{"commentaire sur code", "g-syn-comment", "g-code"},
		{"erreur sur code", "g-syn-error", "g-code"},
	}

	themes := []struct {
		name string
		vars ThemeVars
	}{
		{"dark", paletteDark},
		{"light", paletteLight},
	}

	for _, theme := range themes {
		for _, pair := range pairs {
			fg, ok := theme.vars[pair.fg]
			if !ok {
				t.Errorf("%s : token %s absent de la palette", theme.name, pair.fg)
				continue
			}
			bg, ok := theme.vars[pair.bg]
			if !ok {
				t.Errorf("%s : token %s absent de la palette", theme.name, pair.bg)
				continue
			}
			if got := contrastRatio(fg, bg); got < 4.5 {
				t.Errorf("%s/%s : %s sur %s = %.2f:1, want >= 4.5:1",
					theme.name, pair.name, pair.fg, pair.bg, got)
			}
		}
	}
}

// TestCodeSurfaceDistinctFromWidget : le fond du code doit se détacher de
// la carte qui le contient. Une panne réelle : `.md code`/`.md pre`
// réutilisaient g-widget — le code était invisible. Dans le canon
// « obsidienne » (comme en clair), la surface de code est une couche PLUS
// PROFONDE que la carte : on vérifie l'orientation de la profondeur et un
// écart minimum (le filet var(--g-border) complète la délimitation).
func TestCodeSurfaceDistinctFromWidget(t *testing.T) {
	themes := []struct {
		name string
		vars ThemeVars
	}{
		{"dark", paletteDark},
		{"light", paletteLight},
	}

	for _, theme := range themes {
		code, ok := theme.vars["g-code"]
		if !ok {
			t.Errorf("%s : token g-code absent de la palette", theme.name)
			continue
		}
		widget, ok := theme.vars["g-widget"]
		if !ok {
			t.Errorf("%s : token g-widget absent de la palette", theme.name)
			continue
		}
		lumCode := relativeLuminance(code)
		lumWidget := relativeLuminance(widget)
		if code == widget || lumCode >= lumWidget {
			t.Errorf("%s : g-code (%s) pas plus profond que g-widget (%s), le fond du code se confond avec la carte", theme.name, code, widget)
			continue
		}
		if got := contrastRatio(code, widget); got < 1.03 {
			t.Errorf("%s : contraste code/carte = %.2f:1, want >= 1.03:1 (couches indistinguables)", theme.name, got)
		}
	}
}

// TestPaletteContrastHelper : garde-fou du calcul (noir sur blanc = 21:1,
// une couleur sur elle-même = 1:1) pour que TestPaletteContrast ne passe
// pas par accident.
func TestPaletteContrastHelper(t *testing.T) {
	if got := contrastRatio("#000000", "#ffffff"); math.Abs(got-21) > 0.01 {
		t.Errorf("contraste noir/blanc = %.2f, want 21", got)
	}
	if got := contrastRatio("#7f8c99", "#7f8c99"); math.Abs(got-1) > 0.001 {
		t.Errorf("contraste identique = %.2f, want 1", got)
	}
}
