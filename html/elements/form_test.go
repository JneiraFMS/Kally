package elements

import (
	"bytes"
	"strings"
	"testing"

	"maragu.dev/gomponents"
)

// renderHTML sérialise un composant gomponents en HTML.
func renderHTML(t *testing.T, node gomponents.Node) string {
	t.Helper()
	var buf bytes.Buffer
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render() inattendu: %v", err)
	}
	return buf.String()
}

// TestFormCustomWithActionHasNativeFallback : le formulaire porte action +
// method=post (il fonctionne sans JavaScript) tout en gardant les attributs
// htmx, plus la désactivation du bouton pendant l'échange.
func TestFormCustomWithActionHasNativeFallback(t *testing.T) {
	html := renderHTML(t, FormCustomWithAction(
		[]ChampFormulaire{{Nom: "id", Type: "text", Label: "Identifiant :"}},
		"/kb",
		"Enregistrer",
	))

	for _, want := range []string{
		`action="/kb"`,
		`method="post"`,
		`hx-post="/kb"`,
		`hx-target="#reponse-serveur"`,
		`hx-disabled-elt="find button"`,
		`<button type="submit"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("formulaire sans %q:\n%s", want, html)
		}
	}
}

// TestInputCustomHasStableFieldID : chaque champ reçoit l'id "field-{nom}"
// (cible du focus/contour rouge en cas d'erreur de validation).
func TestInputCustomHasStableFieldID(t *testing.T) {
	tests := []struct {
		name  string
		champ ChampFormulaire
		want  string
	}{
		{name: "input texte", champ: ChampFormulaire{Nom: "id", Type: "text", Label: "ID"}, want: `id="field-id"`},
		{name: "textarea", champ: ChampFormulaire{Nom: "erreur", Type: "textarea", Label: "Err"}, want: `id="field-erreur"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderHTML(t, InputCustom(tt.champ))
			if !strings.Contains(html, tt.want) {
				t.Errorf("InputCustom() sans %q:\n%s", tt.want, html)
			}
		})
	}
}

// TestInputCustomTextareaMono : un textarea Mono est rendu en police mono,
// avec le nombre de lignes demandé (ou 4 par défaut) et redimensionnable
// (pas de resize:none).
func TestInputCustomTextareaMono(t *testing.T) {
	html := renderHTML(t, InputCustom(ChampFormulaire{
		Nom: "erreur", Type: "textarea", Label: "Erreur", Mono: true, Rows: 8,
	}))

	if !strings.Contains(html, `rows="8"`) {
		t.Errorf("textarea sans rows=8:\n%s", html)
	}
	if !strings.Contains(html, "font-mono") {
		t.Errorf("textarea Mono sans classe font-mono:\n%s", html)
	}
	if strings.Contains(html, "resize-none") {
		t.Errorf("le textarea doit rester redimensionnable")
	}

	// Défaut : 4 lignes, pas de mono
	html = renderHTML(t, InputCustom(ChampFormulaire{Nom: "notes", Type: "textarea", Label: "N"}))
	if strings.Contains(html, "font-mono") || !strings.Contains(html, `rows="4"`) {
		t.Errorf("textarea par défaut invalide:\n%s", html)
	}
}

// TestInputCustomContrast : label et placeholder assez contrastés (passe
// accessibility) : label g-base, placeholder g-muted, jamais les niveaux
// délavés g-subdue / g-muted pour ces éléments.
func TestInputCustomContrast(t *testing.T) {
	html := renderHTML(t, InputCustom(ChampFormulaire{Nom: "id", Type: "text", Label: "ID"}))

	if !strings.Contains(html, `<label class="block text-g-base`) {
		t.Errorf("label sans classe contrastée g-base:\n%s", html)
	}
	if strings.Contains(html, `class="block text-g-muted`) {
		t.Errorf("label encore en g-muted (contraste insuffisant)")
	}
	if !strings.Contains(html, "placeholder-g-muted") {
		t.Errorf("placeholder sans classe contrastée g-muted:\n%s", html)
	}
	if strings.Contains(html, "placeholder-g-subdue") {
		t.Errorf("placeholder encore en g-subdue (contraste insuffisant)")
	}
}
