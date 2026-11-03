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
// accessibility) : label g-text, placeholder g-text (g-muted ne donne que
// ~2.5:1 sur le fond des champs, g-subdue est proscrit).
func TestInputCustomContrast(t *testing.T) {
	html := renderHTML(t, InputCustom(ChampFormulaire{Nom: "id", Type: "text", Label: "ID"}))

	if !strings.Contains(html, `class="block text-g-text`) {
		t.Errorf("label sans classe contrastée g-text:\n%s", html)
	}
	if strings.Contains(html, `class="block text-g-muted`) {
		t.Errorf("label encore en g-muted (contraste insuffisant)")
	}
	if !strings.Contains(html, "placeholder-g-text") {
		t.Errorf("placeholder sans classe contrastée g-text:\n%s", html)
	}
	if strings.Contains(html, "placeholder-g-subdue") || strings.Contains(html, "placeholder-g-muted") {
		t.Errorf("placeholder encore en g-subdue/g-muted (contraste insuffisant)")
	}
}

// TestInputCustomLabelBoundToField : le libellé est relié au champ (for/id),
// pour qu'un clic sur le libellé place le focus dans le champ.
func TestInputCustomLabelBoundToField(t *testing.T) {
	tests := []struct {
		name  string
		champ ChampFormulaire
	}{
		{name: "input", champ: ChampFormulaire{Nom: "tags", Type: "text", Label: "Tags"}},
		{name: "textarea", champ: ChampFormulaire{Nom: "notes", Type: "textarea", Label: "Notes"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderHTML(t, InputCustom(tt.champ))
			if !strings.Contains(html, `<label for="field-`+tt.champ.Nom+`"`) {
				t.Errorf("libellé non relié au champ (for manquant):\n%s", html)
			}
			if !strings.Contains(html, `id="field-`+tt.champ.Nom+`"`) {
				t.Errorf("champ sans id:\n%s", html)
			}
		})
	}
}

// fieldTag extrait la balise du contrôle (input ou textarea) du HTML rendu.
func fieldTag(t *testing.T, html string) string {
	t.Helper()
	start := strings.Index(html, "<input")
	if start < 0 {
		start = strings.Index(html, "<textarea")
	}
	if start < 0 {
		t.Fatalf("aucune balise de champ:\n%s", html)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("balise de champ non fermée:\n%s", html)
	}
	return html[start : start+end+1]
}

// TestInputCustomReadOnlySingleClass : un champ en lecture seule porte UN
// seul attribut class (deux attributs class sont ignorés par le parseur, le
// style lecture seule disparaît) et affiche un style distinct des champs
// éditables.
func TestInputCustomReadOnlySingleClass(t *testing.T) {
	tests := []struct {
		name  string
		champ ChampFormulaire
	}{
		{
			name:  "input",
			champ: ChampFormulaire{Nom: "id", Type: "text", Label: "ID", Valeur: "KB-0001", LectureSeule: true},
		},
		{
			name:  "textarea",
			champ: ChampFormulaire{Nom: "notes", Type: "textarea", Label: "Notes", LectureSeule: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderHTML(t, InputCustom(tt.champ))
			tag := fieldTag(t, html)
			if got := strings.Count(tag, "class=\""); got != 1 {
				t.Errorf("attribut class émis %d fois sur le contrôle, want 1:\n%s", got, tag)
			}
			if strings.Contains(tag, "bg-g-block") {
				t.Errorf("champ lecture seule avec le fond des champs éditables:\n%s", tag)
			}
			if !strings.Contains(tag, "cursor-not-allowed") {
				t.Errorf("champ lecture seule sans curseur dédié:\n%s", tag)
			}
			if !strings.Contains(tag, "bg-g-widget") {
				t.Errorf("champ lecture seule sans fond distinct:\n%s", tag)
			}
			if !strings.Contains(tag, `readonly`) {
				t.Errorf("champ lecture seule sans attribut readonly:\n%s", tag)
			}
		})
	}
}
