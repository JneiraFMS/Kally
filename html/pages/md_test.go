package pages

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	kb "grz/internal/domain/kb"
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

func BenchmarkMarkdownConversion(b *testing.B) {
	markdown := strings.Repeat("Performance note with **bold text** and `inline code`.\n\n", 20) +
		"```go\nfunc compute(value int) int { return value * 2 }\n```\n"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mdToNode(markdown)
	}
}

func TestMarkdownConversionConcurrent(t *testing.T) {
	const workers = 16
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var output bytes.Buffer
			if err := mdToNode("**bold**\n\n```go\nfunc main() { fmt.Println(1) }\n```").Render(&output); err != nil {
				errs <- err
				return
			}
			if !strings.Contains(output.String(), "chroma-kd") {
				errs <- fmt.Errorf("concurrent Markdown render missing highlighted output")
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
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
		"color: var(--g-text)",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("styles du bouton copier sans %q", want)
		}
	}
}

// TestMarkdownListsKeepMarkers : le reset preflight de Tailwind supprime les
// puces (list-style: none). Les listes du markdown doivent les retrouver,
// sinon une énumération devient une suite de paragraphes.
func TestMarkdownListsKeepMarkers(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	for _, want := range []string{
		".md ul { list-style: disc;",
		".md ol { list-style: decimal;",
		"padding-left: 1.25rem",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("feuille inline sans %q (listes markdown sans puces)", want)
		}
	}
}

// TestMarkdownHeadingsHierarchy : les titres écrits par l'auteur dans le
// markdown sont plus grands que les titres de section générés par
// l'application (0,875 rem) et chaque niveau se distingue du suivant.
func TestMarkdownHeadingsHierarchy(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	for _, want := range []string{
		".md h1 { font-size: 1.15rem;",
		".md h2 { font-size: 1.05rem;",
		".md h3 { font-size: 0.95rem;",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("hiérarchie des titres markdown incomplète : %q manquant", want)
		}
	}
	if strings.Contains(html, ".md h1, .md h2, .md h3 {\n\tfont-size") {
		t.Errorf("les trois niveaux de titre markdown partagent encore la même taille")
	}
}

// cssRuleBody retourne le corps de la première règle dont le sélecteur
// commence par selector (la feuille inline est embarquée dans la page).
func cssRuleBody(t *testing.T, html, selector string) string {
	t.Helper()
	start := strings.Index(html, selector)
	if start < 0 {
		t.Fatalf("règle %q absente de la feuille inline", selector)
	}
	open := strings.Index(html[start:], "{")
	if open < 0 {
		t.Fatalf("règle %q sans accolade ouvrante", selector)
	}
	bodyStart := start + open + 1
	end := strings.Index(html[bodyStart:], "}")
	if end < 0 {
		t.Fatalf("règle %q non fermée", selector)
	}
	return html[start : bodyStart+end+1]
}

// TestMarkdownCodeSurface : les balises code (inline et blocs fencés) ont une
// surface propre, distincte de la carte qui les contient. Elles réutilisaient
// g-widget — exactement le fond de la carte — le code était donc invisible.
// Le liseré n'est porté que par les blocs : le code inline reste épuré.
func TestMarkdownCodeSurface(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	for _, selector := range []string{".md code", ".md pre"} {
		rule := cssRuleBody(t, html, selector)
		if !strings.Contains(rule, "background: var(--g-code)") {
			t.Errorf("%s sans fond var(--g-code) :\n%s", selector, rule)
		}
		if strings.Contains(rule, "var(--g-widget)") {
			t.Errorf("%s réutilise la surface de la carte :\n%s", selector, rule)
		}
	}

	code := cssRuleBody(t, html, ".md code")
	if strings.Contains(code, "border:") {
		t.Errorf("le code inline ne doit pas porter de liseré :\n%s", code)
	}

	pre := cssRuleBody(t, html, ".md pre")
	if !strings.Contains(pre, "border: 1px solid var(--g-border)") {
		t.Errorf(".md pre sans liseré (délimitation en thème clair) :\n%s", pre)
	}
	if !strings.Contains(pre, "color: var(--g-hi)") {
		t.Errorf("le texte des blocs de code doit être g-hi (g-text échoue sur g-code) :\n%s", pre)
	}
	// Page détail : pas de barre de défilement horizontale, les longues
	// lignes de code et de trace se replient à la place.
	for _, want := range []string{"white-space: pre-wrap;", "overflow-wrap: anywhere"} {
		if !strings.Contains(pre, want) {
			t.Errorf(".md pre sans %q (barre de défilement horizontale) :\n%s", want, pre)
		}
	}
	if strings.Contains(pre, "overflow-x") {
		t.Errorf(".md pre déclare encore un défilement horizontal :\n%s", pre)
	}
	if !strings.Contains(html, ".md { overflow-wrap: anywhere; }") {
		t.Errorf("conteneur .md sans repli des mots très longs (URL) : barre horizontale possible")
	}
	if !strings.Contains(html, ".md pre code { background: none;") {
		t.Errorf("le code dans un bloc pré garde un double fond")
	}
}

// TestMarkdownSyntaxTokens : la coloration syntaxique passe par les tokens
// g-syn-* (calibrés pour 4.5:1 sur g-code), plus par les accents décoratifs
// g-cyan / g-positive / g-text.
func TestMarkdownSyntaxTokens(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	rules := []struct {
		selector string
		want     string
	}{
		{".chroma-k, .chroma-kd", "var(--g-syn-keyword)"},
		{".chroma-s, .chroma-s1", "var(--g-syn-string)"},
		{".chroma-c, .chroma-c1", "var(--g-syn-comment)"},
		{".chroma-err, .chroma-w", "var(--g-syn-error)"},
		{".chroma-nf, .chroma-fm", "var(--g-hi)"},
	}
	for _, rule := range rules {
		body := cssRuleBody(t, html, rule.selector)
		if !strings.Contains(body, "color: "+rule.want) {
			t.Errorf("%s sans color: %s :\n%s", rule.selector, rule.want, body)
		}
	}

	for _, selector := range []string{
		".chroma-k, ", ".chroma-s, ", ".chroma-c, ",
		".chroma-nf, ", ".chroma-nc, ", ".chroma-nv, ",
	} {
		body := cssRuleBody(t, html, selector)
		for _, forbidden := range []string{"var(--g-cyan)", "var(--g-positive)", "var(--g-text)", "var(--g-negative-text)"} {
			if strings.Contains(body, forbidden) {
				t.Errorf("%s utilise encore %s (contraste insuffisant sur g-code) :\n%s", selector, forbidden, body)
			}
		}
	}
}
