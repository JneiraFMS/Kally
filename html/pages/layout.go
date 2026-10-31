package pages

import (
	"grz/html/elements"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var navElements = []elements.NavbarElements{
	{Label: "Home", Href: "/"},
	{Label: "KB", Href: "/kb"},
}

// layout assemble la structure HTML commune (head, scripts, navbar).
func layout(title string, currentPath string, content []gomponents.Node) gomponents.Node {
	return components.HTML5(components.HTML5Props{
		Title:    title,
		Language: "fr",
		Head: []gomponents.Node{
			Script(Src("https://cdn.tailwindcss.com")),
			Script(Src("https://unpkg.com/htmx.org@1.9.10")), // Pour soumettre sans recharger
		},
		Body: []gomponents.Node{
			elements.NavbarCustom(navElements, currentPath),
			gomponents.Group(content),
		},
	})
}
