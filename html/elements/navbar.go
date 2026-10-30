package elements

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type NavbarElements struct {
	Label string
	Href  string
}

func NavbarCustom(elementList []NavbarElements) gomponents.Node {
	var navItems []gomponents.Node

	for _, item := range elementList {
		navItems = append(navItems, Li(
			A(
				Href(item.Href),
				Class("hover:text-blue-600 transition-colors"),
				gomponents.Text(item.Label),
			),
		))
	}

	return Nav(Class("bg-white shadow-md px-6 py-4"),
		Ol(
			Class("flex items-center space-x-6 text-gray-700 font-medium"),
			gomponents.Group(navItems), // <-- On utilise Group ici pour intégrer la slice
		),
	)
}
