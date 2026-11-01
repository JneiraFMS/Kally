package elements

import (
	"strings"
	"testing"

	"maragu.dev/gomponents"
)

// TestBoutonsInvertedTextToken : le texte inversé des boutons primaires
// utilise le token g-ink, pas une valeur hexadécimale en dur ([#002b36]).
func TestBoutonsInvertedTextToken(t *testing.T) {
	tests := []struct {
		name string
		node gomponents.Node
	}{
		{"primaire", BoutonCustom("Valider", true)},
		{"submit", BoutonSubmit("Enregistrer")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderHTML(t, tt.node)
			if strings.Contains(html, "[#002b36]") {
				t.Errorf("hexadécimal en dur dans le bouton %s:\n%s", tt.name, html)
			}
			if !strings.Contains(html, "text-g-ink") {
				t.Errorf("texte inversé sans token g-ink dans le bouton %s:\n%s", tt.name, html)
			}
		})
	}
}
