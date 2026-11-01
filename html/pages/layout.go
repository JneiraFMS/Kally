package pages

import (
	"fmt"

	"grz/html/elements"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var navElements = []elements.NavbarElements{
	{Label: "Accueil", Href: "/"},
	{Label: "Graphe", Href: "/kb/graph"},
}

// glanceBodyStyle : style global du corps (fond Solarized + typo + focus
// clavier visible) et styles markdown des fiches (.md). Les couleurs
// proviennent de la palette Solarized (palette.go), même source que les
// tokens g-* de tailwind.config.js. La feuille des classes utilitaires,
// générée hors-ligne par tools/build-css.sh, est servie à
// /assets/tailwind.css (web/static/tailwind.css embarqué).
var glanceBodyStyle = fmt.Sprintf(`body {
	background-color: %s;
	color: %s;
	font: 14px/1.6 ui-sans-serif, system-ui, sans-serif;
	-webkit-font-smoothing: antialiased;
}
::selection { background: rgba(178, 137, 0, 0.30); }
a:focus-visible, button:focus-visible, input:focus-visible, textarea:focus-visible {
	outline: 2px solid %s;
	outline-offset: 2px;
}
/* Sections markdown des fiches (classe md) : épaisseur réduite, fonds
   blocs solarisés, code mono, liens accentués. */
.md p { margin: 0.25rem 0; }
.md ul, .md ol { margin: 0.25rem 0; padding-left: 1.25rem; }
.md h1, .md h2, .md h3 {
	font-size: 0.9rem; font-weight: 600; color: %s;
	margin: 0.5rem 0 0.25rem;
}
.md code {
	font: 12px ui-monospace, SFMono-Regular, monospace;
	color: %s; background: %s;
	padding: 0.1rem 0.3rem; border-radius: 0.25rem;
}
.md pre {
	background: %s; color: %s;
	border-radius: 0.5rem; padding: 0.75rem 1rem; margin: 0.5rem 0;
	overflow-x: auto;
	font: 12px/1.5 ui-monospace, SFMono-Regular, monospace;
}
.md pre code { background: none; padding: 0; color: inherit; }
.md a { color: %s; text-decoration: underline; }
.md blockquote {
	border-left: 2px solid %s; padding-left: 0.75rem;
	color: %s; margin: 0.25rem 0;
}`,
	solarizedBase03, solarizedBase1, solarizedCyan, // body + focus (outline cyan)
	solarizedBase2,                  // titres .md
	solarizedBase2, solarizedBase02, // .md code (couleur + fond)
	solarizedBase02, solarizedBase1, // .md pre (fond + texte)
	solarizedYellow,               // liens .md
	solarizedSep, solarizedBase00, // blockquote (bord + texte)
)

// layout assemble la structure HTML commune (head, scripts, navbar)
// avec la palette Solarized délavée.
func layout(title string, currentPath string, content []gomponents.Node) gomponents.Node {
	return components.HTML5(components.HTML5Props{
		Title:    title,
		Language: "fr",
		Head: []gomponents.Node{
			// Feuille locale générée hors-ligne (tools/build-css.sh) :
			// classes utilitaires + preflight, palette Solarized en dur.
			Link(Rel("stylesheet"), Href("/assets/tailwind.css")),
			gomponents.El("style", gomponents.Raw(glanceBodyStyle)), // fond + typo du body
			Script(Src("https://unpkg.com/htmx.org@1.9.10")),        // Pour soumettre sans recharger
		},
		Body: []gomponents.Node{
			elements.NavbarCustom(navElements, currentPath),
			gomponents.Group(content),
		},
	})
}
