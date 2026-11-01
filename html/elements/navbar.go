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

// NavbarCustom : barre de navigation minimaliste — wordmark à gauche,
// items de nav à droite, item actif souligné d'un fin trait jaune.
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

	return Nav(Class("bg-g-widget px-6"),
		Div(Class("flex items-center justify-between max-w-5xl mx-auto"),
			// Wordmark en mono
			A(Href("/"),
				Class("py-3 font-mono text-g-hi text-sm font-semibold tracking-wide"),
				gomponents.Text("KB"),
			),
			Ol(
				Class("flex items-center gap-6 text-sm font-medium"),
				gomponents.Group(navItems),
			),
		),
	)
}
