package pages

import (
	"strings"
	"testing"
)

// TestLayoutFontFaces : les deux polices variables sont déclarées via
// @font-face pointant l'embed self-hosted (aucune référence à un CDN de
// polices), et le body les utilise — la stack reste dégradée proprement
// vers les familles système.
func TestLayoutFontFaces(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	for _, want := range []string{
		`@font-face`,
		`font-family: "Plus Jakarta Sans"`,
		`src: url("/assets/fonts/plus-jakarta-sans.woff2")`,
		`font-family: "JetBrains Mono"`,
		`src: url("/assets/fonts/jetbrains-mono.woff2")`,
		`font-display: swap`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("head sans %q (polices self-hosted manquantes)", want)
		}
	}

	if strings.Contains(html, "fonts.googleapis.com") || strings.Contains(html, "fonts.gstatic.com") {
		t.Errorf("référence à un CDN de polices : la typo doit être self-hosted")
	}

	for _, want := range []string{
		`"Plus Jakarta Sans", ui-sans-serif`,
		`"JetBrains Mono"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("stack de police sans %q", want)
		}
	}
}
