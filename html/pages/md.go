package pages

import (
	"bytes"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting/v2"
)

// goldmarkOptions : rendu markdown de GRZ.
//   - highlighting : coloration syntaxique des blocs de code, en MODE
//     CLASSES (chromahtml.WithClasses + préfixe chroma-) : aucune couleur
//     inline émise — le thème actif (dark/light) colore via les var(--g-*)
//     définies dans layout.go (.chroma-*). Langages connus seulement,
//     sans détection auto (langage inconnu → bloc inerte, sûr).
var goldmarkOptions = []goldmark.Option{
	goldmark.WithExtensions(highlighting.NewHighlighting(
		highlighting.WithFormatOptions(
			chromahtml.WithClasses(true),
			chromahtml.ClassPrefix("chroma-"),
		),
	)),
}

// mdToNode rend du markdown en HTML sûr via goldmark : le HTML brut du
// contenu est ignoré (option Unsafe désactivée, comportement par défaut),
// les enrichissements markdown restent fonctionnels (gras, listes, code).
// En cas d'erreur de rendu, le texte brut est affiché (échappé par
// gomponents.Text) plutôt qu'une fiche vide.
func mdToNode(md string) gomponents.Node {
	markdown := goldmark.New(goldmarkOptions...)
	var buf bytes.Buffer
	if err := markdown.Convert([]byte(md), &buf); err != nil {
		return Div(Class("md"), gomponents.Text(md))
	}
	return Div(Class("md"), gomponents.Raw(buf.String()))
}
