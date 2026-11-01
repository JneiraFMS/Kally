package pages

import (
	"strings"
	"testing"

	"grz/html/elements"
	"grz/internal/kb"
)

// TestPageKBCreateFieldsConfiguration : le champ « erreur » du formulaire de
// création est monospace sur 8 lignes ; Notes/Résolution passent à 6 lignes.
func TestPageKBCreateFieldsConfiguration(t *testing.T) {
	errField, _ := findField(kbFormFields, "erreur")
	if errField == nil {
		t.Fatal("champ erreur manquant dans kbFormFields")
	}
	if !errField.Mono || errField.Rows != 8 {
		t.Errorf("champ erreur = Mono=%v Rows=%d, want Mono=true Rows=8", errField.Mono, errField.Rows)
	}
	assertRows(t, kbFormFields, map[string]int{"notes": 6, "resolution": 6})
}

// TestPageKBEditFieldsConfiguration : le formulaire d'édition applique la
// même configuration et la rend en mono/rows=8.
func TestPageKBEditFieldsConfiguration(t *testing.T) {
	k := mustCardKB(t, "KB-0001")

	errField, _ := findField(kbEditFields(k), "erreur")
	if errField == nil {
		t.Fatal("champ erreur manquant dans kbEditFields")
	}
	if !errField.Mono || errField.Rows != 8 {
		t.Errorf("champ erreur (edit) = Mono=%v Rows=%d, want Mono=true Rows=8", errField.Mono, errField.Rows)
	}
	assertRows(t, kbEditFields(k), map[string]int{"notes": 6, "resolution": 6})

	html := renderNode(t, PageKBEdit("/kb/edit", k))
	if !strings.Contains(html, "font-mono") {
		t.Errorf("champ erreur d'édition non rendu en mono:\n%s", html)
	}
	if !strings.Contains(html, `rows="8"`) {
		t.Errorf("champ erreur d'édition sans rows=8:\n%s", html)
	}
}

// TestPageKBCreateMarkdownHint : les libellés signalent que Notes et
// Résolution acceptent le markdown.
func TestPageKBCreateMarkdownHint(t *testing.T) {
	html := renderNode(t, PageKBCreate("/kb/new"))
	if !strings.Contains(html, "Notes (markdown) :") || !strings.Contains(html, "Résolution (markdown) :") {
		t.Errorf("libellés sans indication markdown:\n%s", html)
	}
}

// TestKBEditFieldsValidValues : kbEditFields pré-remplit correctement les
// valeurs depuis la KB (tags regroupés).
func TestKBEditFieldsValidValues(t *testing.T) {
	k, err := kb.NewKB("KB-0002", "err", "notes", "résolution", "calcul", "kpi")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	tags, _ := findField(kbEditFields(k), "tags")
	if tags == nil || tags.Valeur != "calcul kpi" {
		t.Errorf("champ tags = %+v, want valeur \"calcul kpi\"", tags)
	}
}

func findField(fields []elements.ChampFormulaire, nom string) (*elements.ChampFormulaire, int) {
	for i := range fields {
		if fields[i].Nom == nom {
			return &fields[i], i
		}
	}
	return nil, -1
}

func assertRows(t *testing.T, fields []elements.ChampFormulaire, rows map[string]int) {
	t.Helper()
	for nom, want := range rows {
		f, _ := findField(fields, nom)
		if f == nil {
			t.Errorf("champ %s manquant", nom)
			continue
		}
		if f.Rows != want {
			t.Errorf("champ %s Rows=%d, want %d", nom, f.Rows, want)
		}
	}
}
