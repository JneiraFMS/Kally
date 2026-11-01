package pages

import (
	"strings"
	"testing"

	"grz/internal/kb"
)

// markdownDetailHTML retourne le HTML du détail pour un contenu donné.
func markdownDetailHTML(t *testing.T, erreur, notes, resolution string) string {
	t.Helper()
	k, err := kb.NewKB("KB-0001", erreur, notes, resolution)
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	return renderNode(t, PageKBDetail("/kb/detail", k, nil))
}

// TestMarkdownCodeBlocksHighlighted : les blocs de code à langage connu
// sont rendus avec un code Chroma en MODE CLASSES (spans classés), pas de
// couleurs inline — le thème (dark/light) reste maîtrisé côté CSS.
func TestMarkdownCodeBlocksHighlighted(t *testing.T) {
	md := "Avant.\n\n```go\nfunc main() {\n\tx := 42 // commentaire\n\tfmt.Println(\"hi\")\n}\n```\n"
	html := markdownDetailHTML(t, "e", md, "")

	if !strings.Contains(html, `class="chroma`) {
		t.Errorf("bloc de code sans classe chroma (with-classes) :\n%s", html)
	}
	// Spans classés présents (keyword / fonction / commentaire…)
	if strings.Count(html, `<span class="chroma-`) < 3 {
		t.Errorf("pas de tokens classés chroma- (mode classes désactivé ?)")
	}

	// Le lexeur correspond au langage : chroma-kd ("keyword definition")
	// n'apparaît que si le bloc ```go a été analysé par le lexeur Go
	// (langage connu → lexage, sinon aucun span).
	if !strings.Contains(html, `"chroma-kd"`) {
		t.Errorf("bloc ```go sans lexage spécifique (kd attendu pour func)")
	}
}

// TestChromaClassesCovered : les familles de tokens émises par l'exemple Go
// (kd = `func`, nf = nom de fonction, cl/line = structural) ont une règle
// CSS correspondante dans la feuille inline.
func TestChromaClassesCovered(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	for _, rule := range []string{".chroma-kd", ".chroma-nf", ".chroma-c1", ".chroma-s1"} {
		if !strings.Contains(html, rule) {
			t.Errorf("feuille inline sans la règle %s", rule)
		}
	}
}

// TestMarkdownHighlightingNeverInlineColors : aucun style hex inline dans
// le HTML du markdown (l'inverse des options Unsafe/inline-styles de
// Chroma). Seule la zone .md est vérifiée — la navbar peut porter un style
// inline sur son logo SVG (fill: var(--g-accent), lisible côté CSS).
func TestMarkdownHighlightingNeverInlineColors(t *testing.T) {
	md := "```bash\ncurl -s http://x | grep ripgrep # cmd\n```\n\n```json\n{\"cle\": 1}\n```\n"
	html := markdownDetailHTML(t, "e", md, "")

	// Extraire uniquement le contenu markdown (div.md) pour le test
	start := strings.Index(html, `class="md"`)
	if start < 0 {
		t.Fatalf("zone .md introuvable dans le HTML")
	}
	end := strings.Index(html[start:], `</div>`)
	if end < 0 {
		t.Fatalf("fin de zone .md introuvable")
	}
	mdHTML := html[start : start+end]

	// Chroma ne doit pas émettre de style inline (mode classes imposé)
	if strings.Contains(mdHTML, `style="`) {
		start := strings.Index(mdHTML, `style="`)
		t.Errorf("style inline (Chroma sans mode classes) : %s…", mdHTML[start:start+60])
	}
}

// TestMarkdownUnknownLanguageHarmless : un fence avec un langage inconnu ou
// sans langage reste du code inerte stylé, pas de crash ni de classes
// chroma fantaisistes sur du texte libre.
func TestMarkdownUnknownLanguageHarmless(t *testing.T) {
	html := markdownDetailHTML(t, "e", "```langagebidon\nquoi: que\n```\n", "")

	if strings.Contains(html, `<span class="chroma-`) {
		t.Errorf("langue inconnue colorée à tort")
	}
	if !strings.Contains(html, "<code") {
		t.Errorf("bloc non marqué code")
	}
}
