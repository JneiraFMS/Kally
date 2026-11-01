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
	tests := []struct {
		name string
		md   string
	}{
		{name: "langage inconnu", md: "```langagebidon\nquoi: que\n```\n"},
		{name: "sans langage", md: "```\nquoi: que\n```\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := markdownDetailHTML(t, "e", tt.md, "")
			if strings.Contains(html, `<span class="chroma-`) {
				t.Errorf("langue inconnue colorée à tort")
			}
			if !strings.Contains(html, "<code") {
				t.Errorf("bloc non marqué code")
			}
			if !strings.Contains(html, `class="md-code-copy"`) {
				t.Errorf("bloc non highlighté sans bouton copier")
			}
		})
	}
}

// TestMarkdownCodeBlocksHaveCopyButton : seuls les blocs fenced reçoivent un
// bouton copier ; le code inline reste du texte en ligne sans contrôle ajouté.
func TestMarkdownCodeBlocksHaveCopyButton(t *testing.T) {
	md := "inline `copier()`\n\n```go\nfmt.Println(\"copier\")\n```\n"
	html := markdownDetailHTML(t, "e", md, "")

	if got := strings.Count(html, `class="md-code-copy"`); got != 1 {
		t.Errorf("nombre de boutons copier = %d, want 1 (un bloc fenced, code inline ignoré)\n%s", got, html)
	}
	blockAt := strings.Index(html, `class="md-code-block"`)
	if blockAt < 0 {
		t.Fatal("wrapper du bloc de code manquant")
	}
	block := html[blockAt:]
	buttonAt := strings.Index(block, `class="md-code-copy"`)
	preAt := strings.Index(block, "<pre")
	if buttonAt < 0 || preAt < 0 || buttonAt > preAt {
		t.Errorf("le bouton copier doit précéder le <pre> du bloc de code")
	}
	for _, want := range []string{`aria-label="Copier le code"`, `aria-live="polite"`, `class="md-code-block"`} {
		if !strings.Contains(html, want) {
			t.Errorf("bloc de code sans %q", want)
		}
	}
	if !strings.Contains(block, `<svg viewBox="0 0 24 24"`) || !strings.Contains(block, `aria-hidden="true"`) {
		t.Errorf("le bouton Copier doit montrer une icône SVG décorative")
	}
	if strings.Contains(block, `title="Copier le code">Copier</button>`) {
		t.Errorf("le libellé textuel Copier est encore affiché dans le bouton")
	}
	if !strings.Contains(html, "inline <code>copier()</code>") {
		t.Errorf("le code inline n'est pas préservé")
	}
}

// TestMarkdownCodeCopyScript : le bouton copie le texte du code (textContent,
// donc sans les spans Chroma), préfère Clipboard API et a un fallback navigateur.
func TestMarkdownCodeCopyScript(t *testing.T) {
	html := markdownDetailHTML(t, "e", "```go\nfmt.Println(1)\n```\n", "")

	for _, want := range []string{
		"md-code-copy-script",
		`querySelector("pre > code")`,
		"code.textContent",
		"navigator.clipboard.writeText",
		`document.execCommand("copy")`,
		"copyIcon",
		"copiedIcon",
		"copyErrorIcon",
		"Code copié dans le presse-papiers",
		"Impossible de copier le code",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("script copier sans %q", want)
		}
	}
}

// TestMarkdownCodeCopyThemeStyles : le bouton et son espace réservé utilisent
// les tokens Solarized, de sorte que le contrôle suive le thème actif.
func TestMarkdownCodeCopyThemeStyles(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))
	for _, want := range []string{
		".md-code-block",
		".md-code-copy",
		"padding-top: 2.5rem",
		"background: var(--g-block)",
		"color: var(--g-base)",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("styles du bouton copier sans %q", want)
		}
	}
}
