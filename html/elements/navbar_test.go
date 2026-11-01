package elements

import (
	"strings"
	"testing"
)

// TestNavbarCustomRendersLogo : le wordmark texte « KB » est remplacé par un
// logo SVG lié à l'accueil, avec nom accessible (aria-label), sans texte
// « KB » nu et avec le nœud central thématisé (var(--g-accent)).
func TestNavbarCustomRendersLogo(t *testing.T) {
	html := renderHTML(t, NavbarCustom([]NavbarElements{{Label: "Accueil", Href: "/"}}, "/"))

	if !strings.Contains(html, "<svg") {
		t.Errorf("navbar sans logo SVG:\n%s", html)
	}
	if !strings.Contains(html, `href="/"`) {
		t.Errorf("logo sans lien vers l'accueil")
	}
	if !strings.Contains(html, `aria-label="Accueil"`) {
		t.Errorf("lien-logo sans nom accessible")
	}
	if strings.Contains(html, ">KB</a>") {
		t.Errorf("wordmark texte « KB » toujours présent")
	}
	// Nœud central accenté — couleur via la variable du thème (jamais un
	// attribut fill avec var(), invalide en SVG : style= ou classe).
	if !strings.Contains(html, "var(--g-accent)") {
		t.Errorf("nœud central du logo sans accent thématisé")
	}
	// Logo entièrement decoratif côté SVG (aria-hidden), nom porté par le lien
	if !strings.Contains(html, `aria-hidden="true"`) {
		t.Errorf("logo SVG sans aria-hidden")
	}
}
