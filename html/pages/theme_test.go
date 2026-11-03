package pages

import (
	"strings"
	"testing"
)

// themeInitScriptMarker : le script d'initialisation du thème (anti-FOUC)
// porte ce marqueur et doit précéder le chargement de la feuille CSS.
const themeInitScriptMarker = "grz-theme-init"

// TestLayoutThemeVarRules : le head porte `:root` (dark, par défaut) et
// `[data-theme="light"]` (mode clair) avec toutes les valeurs hex des deux
// palettes ; les styles du body utilisent les var(--g-*).
func TestLayoutThemeVarRules(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	rootStart := strings.Index(html, ":root{")
	lightStart := strings.Index(html, `[data-theme="light"]{`)
	if rootStart < 0 || lightStart < 0 {
		t.Fatalf("règles de thème manquantes (root=%d light=%d)", rootStart, lightStart)
	}
	rootCSS := html[rootStart : strings.Index(html[lightStart:], "}")+lightStart]

	for name, hex := range paletteDark {
		decl := "--" + name + ":" + hex
		if !strings.Contains(rootCSS, decl) {
			t.Errorf(":root sans %s", decl)
		}
	}
	lightCSS := html[lightStart : strings.Index(html[lightStart:], "}")+lightStart]
	for name, hex := range paletteLight {
		decl := "--" + name + ":" + hex
		if !strings.Contains(lightCSS, decl) {
			t.Errorf("[data-theme=light] sans %s", decl)
		}
	}

	// Color-scheme natif (scrollbars, form controls)
	if !strings.Contains(rootCSS, "color-scheme: dark") {
		t.Errorf(":root sans color-scheme: dark")
	}
	if !strings.Contains(lightCSS, "color-scheme: light") {
		t.Errorf("thème light sans color-scheme: light")
	}

	// Le style body consomme les variables (plus d'hex dans le body)
	if !strings.Contains(html, "background-color: var(--g-bg)") {
		t.Errorf("body : fond non lié à var(--g-bg)")
	}
	if !strings.Contains(html, `color: var(--g-text)`) {
		t.Errorf("body : texte non lié à var(--g-text)")
	}
}

// TestLayoutThemeInitScriptBeforeCSS : le script d'init anti-FOUC applique
// data-theme AVANT le chargement de la feuille, et respecte la préférence
// système en premier usage.
func TestLayoutThemeInitScriptBeforeCSS(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	marker := strings.Index(html, themeInitScriptMarker)
	link := strings.Index(html, "/assets/tailwind.css")
	if marker < 0 {
		t.Fatalf("script d'init du thème absent (marqueur %q)", themeInitScriptMarker)
	}
	if link < 0 {
		t.Fatalf("lien CSS absent")
	}
	if marker > link {
		t.Errorf("script d'init (%d) après le lien CSS (%d) : FOUC garanti", marker, link)
	}
	for _, frag := range []string{"localStorage", "prefers-color-scheme"} {
		if !strings.Contains(html, frag) {
			t.Errorf("script d'init sans %s", frag)
		}
	}
}

// TestLayoutThemeToggle : la navbar expose un bouton bascule (identification
// stable, accessible), câblé à un script de bascule.
func TestLayoutThemeToggle(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	for _, want := range []string{
		`id="theme-toggle"`,
		`aria-label="Changer de thème"`,
		"theme-toggle", // handlers câblés au même id
	} {
		if !strings.Contains(html, want) {
			t.Errorf("élément/sélecteur %q manquant", want)
		}
	}
	if strings.Count(html, "theme-toggle") < 2 {
		t.Errorf("bouton présent mais script de bascule manquant")
	}
}

// TestThemeToggleHitArea : la règle #theme-toggle ne remet pas le padding à
// zéro (elle annulait la classe py-3 du bouton) ; la cible de clic doit
// rester au moins aussi haute que le reste de la navbar.
func TestThemeToggleHitArea(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	start := strings.Index(html, "#theme-toggle {")
	if start < 0 {
		t.Fatal("règle #theme-toggle absente")
	}
	end := strings.Index(html[start:], "}")
	if end < 0 {
		t.Fatal("règle #theme-toggle non fermée")
	}
	rule := html[start : start+end]
	if strings.Contains(rule, "padding: 0") {
		t.Errorf("#theme-toggle remet le padding à zéro (zone cliquable écrasée):\n%s", rule)
	}
	if !strings.Contains(html, `class="px-2 py-3"`) {
		t.Errorf("bouton de thème sans padding utilitaire (cible de clic trop petite)")
	}
}
