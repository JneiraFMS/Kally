package pages

import (
	"strings"
	"testing"
)

// tableMD est un tableau pipes GitHub avec colonne centrée et un lien wiki
// dans une cellule.
const tableMD = "| Étape | Statut | [[KB-0002]] |\n" +
	"|:-----:| -------| ----------- |\n" +
	"| un    | ok     | x           |\n" +
	"| deux  | raté   | y           |\n\n" +
	"après.\n"

// TestMarkdownTablesRendered : l'extension Table goldmark est activée — les
// tableaux pipes GitHub deviennent des tables HTML (sinon texte brut), avec
// l'alignement demandé et les liens wiki rendus dans les cellules.
func TestMarkdownTablesRendered(t *testing.T) {
	html := markdownDetailHTML(t, "e", tableMD, "")

	for _, want := range []string{"<table>", "<thead>", "<th>", "<tbody>", "<td>"} {
		if !strings.Contains(html, want) {
			t.Errorf("tableaux markdown sans %q :\n%s", want, html)
		}
	}
	// Alignement centré demandé (|:---:|) : goldmark le pose en style inline
	// (le seul cas .md autorisé), sans espace autour de ':'. Le test vérifie
	// la valeur certifiant l'urbanisme, pas l'espace.
	if !strings.Contains(html, "text-align:center") {
		t.Errorf("alignement centré du tableau non rendu :\n%s", html)
	}
	// Le lien wiki de l'en-tête reste un lien vers la fiche.
	if !strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("lien wiki ignoré dans une cellule de tableau :\n%s", html)
	}
}

// TestMarkdownTableStyles : les tables restent lisibles dans la fiche —
// bordures fines thémées, en-tête sur fond bloc, et un tableau large défile
// horizontalement au lieu de déborder du cadre.
func TestMarkdownTableStyles(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	tableRule := cssRuleBody(t, html, ".md table")
	for _, want := range []string{"border-collapse: collapse", "overflow-x: auto", "max-width: 100%"} {
		if !strings.Contains(tableRule, want) {
			t.Errorf(".md table sans %q :\n%s", want, tableRule)
		}
	}

	// Le liseré portrait-attribut vit dans la règle combinée
	// ".md th, .md td" (précédente), pas dans les spécialisations.
	rule := cssRuleBody(t, html, ".md th, .md td")
	if !strings.Contains(rule, "border: 1px solid var(--g-border)") {
		t.Errorf("cellules sans bordure var(--g-border) :\n%s", rule)
	}

	// ".md th {" (avec accolade) évite d'attraper la règle combinée
	// ".md th, .md td" — les règles combinées précèdent la spécialisation.
	header := cssRuleBody(t, html, ".md th {")
	if !strings.Contains(header, "background: var(--g-block)") {
		t.Errorf(".md th sans fond bloc :\n%s", header)
	}
	if !strings.Contains(header, "color: var(--g-hi)") {
		t.Errorf(".md th sans texte accentué :\n%s", header)
	}
}
