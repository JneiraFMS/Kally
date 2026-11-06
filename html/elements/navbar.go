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

// NavbarPrimary est l'action principale de la navbar (CTA, zone droite).
// Zéro ⇔ pas de CTA rendu.
type NavbarPrimary struct {
	Label string
	Href  string
}

// logoSVG est la marque Kally : un petit graphe de KB liées — nœud central
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

// NavbarCustom : barre de navigation à trois zones sur une seule ligne
// (h-16), adossée à un filet bas fin :
//  1. marque    — constellation SVG liée à l'accueil, wordmark « Kally »
//     et sous-titre monospace miniature « KNOWLEDGE BASE » ;
//  2. navigation — liens textuels (text-sm font-medium), item actif
//     souligné d'un filet fin thématisé ;
//  3. actions    — bascule clair/sombre (comportement scripté par le
//     layout) et CTA primaire plein accent (NavbarPrimary zéro ⇔ pas de
//     CTA).
func NavbarCustom(elementList []NavbarElements, currentPath string, primary NavbarPrimary) gomponents.Node {
	var navItems []gomponents.Node

	for _, item := range elementList {
		isActive := item.Href == currentPath
		linkClass := "relative py-2 text-g-text hover:text-g-hi transition-colors text-sm font-medium"
		if isActive {
			// Soulignement fin thématisé sous l'item actif
			linkClass = "relative py-2 text-g-hi font-medium text-sm after:absolute after:inset-x-0 after:bottom-0 after:h-0.5 after:bg-g-accent after:content-['']"
		}

		navItems = append(navItems, A(
			Href(item.Href),
			Class(linkClass),
			gomponents.Text(item.Label),
		))
	}

	// Zone 3 : bascule de thème + CTA primaire
	actions := []gomponents.Node{
		Button(
			Type("button"),
			ID("theme-toggle"),
			gomponents.Attr("aria-label", "Changer de thème"),
			Class("px-2 py-3"),
			gomponents.Text("◐"),
		),
	}
	if primary.Label != "" && primary.Href != "" {
		actions = append(actions,
			A(Href(primary.Href),
				Class("hidden sm:inline-block px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 text-xs font-semibold shadow-sm transition"),
				gomponents.Text(primary.Label),
			),
		)
	}

	return Nav(
		Class("w-full border-b border-g-border bg-g-widget"),
		Div(Class("max-w-5xl mx-auto px-6 h-16 flex items-center justify-between"),
			// Zone 1 : marque — constellation + wordmark + sous-titre mono,
			// regroupés en un seul bloc (le row distribue exactement 3 zones).
			// La cellule logo garde sa largeur fixe : sa position ne bouge
			// jamais quelle que soit la page.
			Div(Class("flex items-center gap-2.5"),
				Div(Class("flex items-center w-10 shrink-0"),
					A(Href("/"),
						Class("flex items-center text-g-hi hover:opacity-80 transition-opacity"),
						gomponents.Attr("aria-label", "Accueil"),
						logoSVG(),
					),
				),
				Span(Class("block text-sm font-semibold leading-tight text-g-hi"),
					gomponents.Text("Kally"),
				),
				Span(Class("hidden sm:block font-mono text-[9px] uppercase tracking-[0.2em] leading-tight text-g-text"),
					gomponents.Text("KNOWLEDGE BASE"),
				),
			),
			// Zone 2 : navigation textuelle
			Ol(
				Class("flex items-center gap-6 text-sm font-medium list-none"),
				gomponents.Group(navItems),
			),
			// Zone 3 : actions (thème + CTA)
			Div(Class("flex items-center gap-3"), gomponents.Group(actions)),
		),
	)
}
