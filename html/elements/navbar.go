package elements

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// NavbarElements est un item de la barre de navigation.
type NavbarElements struct {
	Label string
	Href  string
}

// logoSVG est la marque GRZ : un petit graphe de KB liées — nœud central
// accentué (g-accent, suit le thème), satellites et arêtes en currentColor
// (la couleur du texte du lien qui l'enveloppe).
func logoSVG() gomponents.Node {
	return gomponents.El("svg",
		gomponents.Attr("viewBox", "0 0 24 24"),
		gomponents.Attr("width", "20"),
		gomponents.Attr("height", "20"),
		gomponents.Attr("aria-hidden", "true"),
		gomponents.Attr("fill", "none"),
		// Arêtes (current color + opacité pour rester discrètes)
		gomponents.El("path",
			gomponents.Attr("d", "M12 12L5 6.5M12 12L19 7.5M12 12L13.5 19.5"),
			gomponents.Attr("stroke", "currentColor"),
			gomponents.Attr("stroke-width", "1.5"),
			gomponents.Attr("stroke-linecap", "round"),
			gomponents.Attr("opacity", "0.4"),
		),
		gomponents.El("circle",
			gomponents.Attr("cx", "5"), gomponents.Attr("cy", "6.5"), gomponents.Attr("r", "1.75"),
			gomponents.Attr("fill", "currentColor"), gomponents.Attr("opacity", "0.55"),
		),
		gomponents.El("circle",
			gomponents.Attr("cx", "19"), gomponents.Attr("cy", "7.5"), gomponents.Attr("r", "1.75"),
			gomponents.Attr("fill", "currentColor"), gomponents.Attr("opacity", "0.55"),
		),
		gomponents.El("circle",
			gomponents.Attr("cx", "13.5"), gomponents.Attr("cy", "19.5"), gomponents.Attr("r", "1.75"),
			gomponents.Attr("fill", "currentColor"), gomponents.Attr("opacity", "0.55"),
		),
		// Nœud central — couleur via la variable du thème (presentation
		// attributes SVG n'acceptent pas var() : style= est le bon endroit)
		gomponents.El("circle",
			gomponents.Attr("cx", "12"), gomponents.Attr("cy", "12"), gomponents.Attr("r", "3.25"),
			gomponents.Attr("style", "fill: var(--g-accent)"),
		),
	)
}

// NavbarCustom : barre de navigation minimaliste — logo à gauche (vers
// l'accueil), items de nav au centre, bascule clair/sombre à droite,
// item actif souligné d'un fin trait jaune.
func NavbarCustom(elementList []NavbarElements, currentPath string) gomponents.Node {
	var navItems []gomponents.Node

	for _, item := range elementList {
		isActive := item.Href == currentPath
		linkClass := "relative py-3 text-g-base hover:text-g-hi transition-colors"
		if isActive {
			// Soulignement fin jaune sous l'item actif
			linkClass = "relative py-3 text-g-hi font-medium after:absolute after:inset-x-0 after:bottom-0 after:h-0.5 after:bg-g-accent after:content-['']"
		}

		navItems = append(navItems, Li(
			A(
				Href(item.Href),
				Class(linkClass),
				gomponents.Text(item.Label),
			),
		))
	}

	return Nav(Class("bg-g-widget"),
		Div(Class("max-w-5xl mx-auto px-6 flex items-center justify-between"),
			// Logo GRZ — cellule de largeur fixe pour que sa position ne
			// bouge jamais quelle que soit la page
			Div(Class("flex items-center w-10 shrink-0"),
				A(Href("/"),
					Class("py-3 flex items-center text-g-hi hover:opacity-80 transition-opacity"),
					gomponents.Attr("aria-label", "Accueil"),
					logoSVG(),
				),
			),
			Div(Class("flex items-center gap-6"),
				Ol(
					Class("flex items-center gap-6 text-sm font-medium list-none"),
					gomponents.Group(navItems),
				),
				// Bascule clair/sombre (comportement : themeToggleScript)
				Button(
					Type("button"),
					ID("theme-toggle"),
					gomponents.Attr("aria-label", "Changer de thème"),
					Class("py-3"),
					gomponents.Text("◐"),
				),
			),
		),
	)
}
