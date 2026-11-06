package pages

import (
	"strings"
	"testing"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"
)

// textareaTag retourne la balise ouvrante du textarea portant l'id du champ,
// pour vérifier ses attributs sans confondre les champs entre eux.
func textareaTag(t *testing.T, html, field string) string {
	t.Helper()
	start := strings.Index(html, `id="field-`+field+`"`)
	if start < 0 {
		t.Fatalf("champ %q absent du formulaire", field)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("balise du champ %q non terminée", field)
	}
	return html[start : start+end]
}

// TestKBFormAcceptsImagePaste : Notes et Résolution sont marqués pour le
// collage d'images ; Erreur (trace brute) reste hors périmètre.
func TestKBFormAcceptsImagePaste(t *testing.T) {
	html := renderNode(t, PageKBCreate("/kb/new"))

	for _, field := range []string{"notes", "resolution"} {
		if tag := textareaTag(t, html, field); !strings.Contains(tag, "data-image-paste") {
			t.Errorf("textarea %q sans data-image-paste: %s", field, tag)
		}
	}
	if tag := textareaTag(t, html, "erreur"); strings.Contains(tag, "data-image-paste") {
		t.Errorf("textarea erreur ne doit pas accepter d'image: %s", tag)
	}
	if !strings.Contains(html, "grz-image-paste") {
		t.Error("script de collage absent du formulaire KB")
	}
}

// TestJournalFormAcceptsImagePaste : le contenu du journal accepte aussi les
// images, en création comme en édition.
func TestJournalFormAcceptsImagePaste(t *testing.T) {
	create := renderNode(t, PageJournalCreateWith("/journal/new", FormulaireJournal{}))
	if tag := textareaTag(t, create, "content"); !strings.Contains(tag, "data-image-paste") {
		t.Errorf("textarea content sans data-image-paste: %s", tag)
	}
	if !strings.Contains(create, "grz-image-paste") {
		t.Error("script de collage absent du formulaire journal")
	}

	edit := renderNode(t, PageJournalEditWith("/journal/edit", "2026-11-05", FormulaireJournal{}))
	if tag := textareaTag(t, edit, "content"); !strings.Contains(tag, "data-image-paste") {
		t.Errorf("textarea content (édition) sans data-image-paste: %s", tag)
	}
}

// TestKBEditFormAcceptsImagePaste : le formulaire d'édition de KB partage la
// même prise en charge des images que la création (la liste de champs d'édition
// avait dérivé et perdait l'option).
func TestKBEditFormAcceptsImagePaste(t *testing.T) {
	k, err := kb.NewKB("KB-0002", "err", "notes", "résolution", "calcul")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	html := renderNode(t, PageKBEdit("/kb/edit", k))

	for _, field := range []string{"notes", "resolution"} {
		if tag := textareaTag(t, html, field); !strings.Contains(tag, "data-image-paste") {
			t.Errorf("textarea %q (édition) sans data-image-paste: %s", field, tag)
		}
	}
	if tag := textareaTag(t, html, "erreur"); strings.Contains(tag, "data-image-paste") {
		t.Errorf("textarea erreur (édition) ne doit pas accepter d'image: %s", tag)
	}
	if !strings.Contains(html, "grz-image-paste") {
		t.Error("script de collage absent du formulaire d'édition")
	}
}

// TestReadPagesDoNotLoadImagePaste : le script de collage ne pollue pas les
// pages de lecture.
func TestReadPagesDoNotLoadImagePaste(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "n", "r")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))
	if strings.Contains(html, "grz-image-paste") {
		t.Error("script de collage présent sur la fiche de lecture")
	}
}

// TestMarkdownRendersPastedImage : une image insérée par le script est bien
// rendue par goldmark, avec un chemin /media/.
func TestMarkdownRendersPastedImage(t *testing.T) {
	html := renderNode(t, mdToNode("![image](/media/abc.png)"))

	if !strings.Contains(html, `<img src="/media/abc.png"`) {
		t.Errorf("image markdown non rendue:\n%s", html)
	}
}

// TestMarkdownImageStyle : les images restent dans la largeur de la fiche.
func TestMarkdownImageStyle(t *testing.T) {
	html := renderNode(t, PageKBCreate("/kb/new"))

	if !strings.Contains(html, ".md img") || !strings.Contains(html, "max-width: 100%") {
		t.Error("règle CSS .md img absente du layout")
	}
}
