package pages

import (
	"bytes"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	"github.com/yuin/goldmark"
)

// mdToNode rend du markdown en HTML sûr via goldmark : le HTML brut du
// contenu est ignoré (option Unsafe désactivée, comportement par défaut),
// les enrichissements markdown restent fonctionnels (gras, listes, code).
// En cas d'erreur de rendu, le texte brut est affiché (échappé par
// gomponents.Text) plutôt qu'une fiche vide.
func mdToNode(md string) gomponents.Node {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return Div(Class("md"), gomponents.Text(md))
	}
	return Div(Class("md"), gomponents.Raw(buf.String()))
}
