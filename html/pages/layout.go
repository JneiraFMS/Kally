package pages

import (
	"grz/html/elements"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var navElements = []elements.NavbarElements{
	{Label: "Accueil", Href: "/"},
	{Label: "Graphe", Href: "/kb/graph"},
}

// glancePalette = palette Solarized (délavée), complète, en dark :
//   - base03 #002b36 (fond), base02 #073642 (widgets)
//   - teintes de texte base1 #93a1a1 ... base01 #586e75 (dé lavées)
//   - accent jaune solarisé délavé #b58900, verts/bleus former
const glancePalette = `tailwind.config = {
	theme: {
		extend: {
			colors: {
				'g-bg': '#002b36',        // base03
				'g-widget': '#073642',    // base02
				'g-block': '#15414d',     // entre 02 et 01 (inputs, blocs)
				'g-hover': '#214c56',
				'g-border': '#1b444f',
				'g-sep': '#28515a',
				'g-accent': '#b58900',    // yellow (délavé)
				'g-accent-dim': '#0c464f',
				'g-hi': '#eee8d5',        // base2 (titres)
				'g-text': '#93a1a1',      // base1
				'g-base': '#839496',      // base0
				'g-muted': '#657b83',     // base00
				'g-subdue': '#586e75',    // base01
				'g-negative': '#dc322f',  // red
				'g-positive': '#859900',  // green
			},
		},
	},
}`

// glanceBodyStyle : style global du corps (fond Solarized, typo 14px,
// focus clavier visible partout — accessibilité).
const glanceBodyStyle = `body {
	background-color: #002b36;
	color: #93a1a1;
	font: 14px/1.6 ui-sans-serif, system-ui, sans-serif;
	-webkit-font-smoothing: antialiased;
}
::selection { background: rgba(178, 137, 0, 0.30); }
a:focus-visible, button:focus-visible, input:focus-visible, textarea:focus-visible {
	outline: 2px solid #2aa198;
	outline-offset: 2px;
}`

// layout assemble la structure HTML commune (head, scripts, navbar)
// avec la palette Solarized délavée.
func layout(title string, currentPath string, content []gomponents.Node) gomponents.Node {
	return components.HTML5(components.HTML5Props{
		Title:    title,
		Language: "fr",
		Head: []gomponents.Node{
			Script(Src("https://cdn.tailwindcss.com")),
			Script(gomponents.Raw(glancePalette)), // palette Glance
			// NB: html.Style() est un ATTRIBUT style="..." : pour une feuille
			// réelle, construire l'élément via El("style", Raw(...)).
			gomponents.El("style", gomponents.Raw(glanceBodyStyle)), // fond + typo du body
			Script(Src("https://unpkg.com/htmx.org@1.9.10")),        // Pour soumettre sans recharger
		},
		Body: []gomponents.Node{
			elements.NavbarCustom(navElements, currentPath),
			gomponents.Group(content),
		},
	})
}
