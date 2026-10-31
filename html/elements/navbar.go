package elements

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type NavbarElements struct {
	Label string
	Href  string
}

func NavbarCustom(elementList []NavbarElements, currentPath string) gomponents.Node {
	var navItems []gomponents.Node

	for _, item := range elementList {

		isActive := item.Href == currentPath
		linkClass := "hover:text-blue-600 transition-colors"
		if isActive {
			linkClass = "text-blue-600 font-bold"
		}

		navItems = append(navItems, Li(
			A(
				Href(item.Href),
				Class(linkClass),
				gomponents.Text(item.Label),
			),
		))
	}

	return Nav(Class("bg-white shadow-md px-6 py-4"),
		Ol(
			Class("flex items-center space-x-6 text-gray-700 font-medium"),
			gomponents.Group(navItems),
		),
	)
}
