package elements

import (
	"strings"
	"testing"
)

// navbar.HTML : rendu de la navbar avec le CTA primaire du layout.
func navbarHTML(t *testing.T) string {
	t.Helper()
	return renderHTML(t, NavbarCustom(
		[]NavbarElements{
			{Label: "Accueil", Href: "/"},
			{Label: "Journal", Href: "/journal"},
			{Label: "Graphe", Href: "/kb/graph"},
		},
		"/",
		NavbarPrimary{Label: "Créer une KB", Href: "/kb/new"},
	))
}

// TestNavbarCustomThreeZones : la navbar est une barre 3 zones sur une
// seule ligne (h-16) : marque à gauche, navigation au centre, actions à
// droite (thème + CTA), le tout délimité par un filet bas fin.
func TestNavbarCustomThreeZones(t *testing.T) {
	html := navbarHTML(t)

	for _, want := range []string{"h-16", "items-center", "justify-between"} {
		if !strings.Contains(html, want) {
			t.Errorf("navbar sans classe %q (structure 3 zones sur une ligne)", want)
		}
	}
	if !strings.Contains(html, "w-full") && !strings.Contains(html, "mx-auto") {
		t.Errorf("navbar non pleine largeur avec conteneur centré")
	}
}

// TestNavbarCustomBrand : zone marque = logo constellation + wordmark
// « Kally » + sous-titre monospace miniature « KNOWLEDGE BASE ».
func TestNavbarCustomBrand(t *testing.T) {
	html := navbarHTML(t)

	if !strings.Contains(html, ">Kally</") {
		t.Errorf("wordmark « Kally » absent dans la zone marque:\n%s", html)
	}
	if !strings.Contains(html, "KNOWLEDGE BASE") {
		t.Errorf("sous-titre « KNOWLEDGE BASE » absent")
	}
	for _, want := range []string{"uppercase", "tracking-[0.2em]", "font-mono"} {
		if !strings.Contains(html, want) {
			t.Errorf("sous-titre de marque sans classe %q", want)
		}
	}
}

// TestNavbarCustomLogoLink : le lien logo reste signé (aria-label, aria-
// hidden SVG, accueil) et garde sa cellule de largeur fixe — sa position
// ne bouge jamais d'une page à l'autre.
func TestNavbarCustomLogoLink(t *testing.T) {
	html := navbarHTML(t)

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
	if !strings.Contains(html, "var(--g-accent)") {
		t.Errorf("nœud central du logo sans accent thématisé")
	}
	if !strings.Contains(html, `aria-hidden="true"`) {
		t.Errorf("logo SVG sans aria-hidden")
	}
	if !strings.Contains(html, `w-10 shrink-0`) {
		t.Errorf("logo sans conteneur de largeur fixe (w-10 shrink-0)")
	}
}

// TestNavbarCustomNavLinks : la navigation est textuelle, en fonte —
// item actif = filet fin thématisé sous le lien.
func TestNavbarCustomNavLinks(t *testing.T) {
	html := navbarHTML(t)

	if !strings.Contains(html, "Accueil") || !strings.Contains(html, "Journal") {
		t.Errorf("liens de navigation absents:\n%s", html)
	}
	// Lien actif (Accueil sur "/") : filet fin soulignant
	if !strings.Contains(html, "after:bg-g-accent") {
		t.Errorf("item actif sans soulignement fin (after:bg-g-accent)")
	}
	if !strings.Contains(html, `href="/journal"`) {
		t.Errorf("lien Journal sans href")
	}
}

// TestNavbarCustomActions : zone actions = bascule de thème (id stable
// câblé par le script du layout) + CTA primaire plein accent, libellé et
// cible injectés par le layout.
func TestNavbarCustomActions(t *testing.T) {
	html := navbarHTML(t)

	if !strings.Contains(html, `id="theme-toggle"`) {
		t.Errorf("bascule de thème absente (id theme-toggle)")
	}
	if !strings.Contains(html, `aria-label="Changer de thème"`) {
		t.Errorf("bascule de thème sans nom accessible")
	}
	if !strings.Contains(html, `href="/kb/new"`) {
		t.Errorf("CTA sans lien")
	}
	if !strings.Contains(html, "Créer une KB") {
		t.Errorf("CTA sans libellé")
	}
	if !strings.Contains(html, "bg-g-accent") {
		t.Errorf("CTA sans fond accent (bg-g-accent)")
	}
}

// TestNavbarCustomNoCTA : un NavbarPrimary zéro ne rend aucun CTA — la
// navbar reste utilisable sans action de création (pages annexes).
func TestNavbarCustomNoCTA(t *testing.T) {
	html := renderHTML(t, NavbarCustom([]NavbarElements{{Label: "Accueil", Href: "/"}}, "/", NavbarPrimary{}))

	if strings.Contains(html, "Créer une KB") {
		t.Errorf("CTA rendu alors que NavbarPrimary est vide")
	}
}
