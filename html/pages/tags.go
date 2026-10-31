package pages

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// tagBadges rend les tags d'une KB en badges cliquables pointant vers le
// filtre /kb?tag=xxx. Retourne une chaîne vide pour une KB sans tag.
func tagBadges(tags []string) gomponents.Node {
	var badges []gomponents.Node
	for _, tag := range tags {
		badges = append(badges,
			A(Href("/kb?tag="+tag),
				Class("inline-block px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 text-xs font-medium hover:bg-blue-100"),
				gomponents.Text(tag),
			),
		)
	}

	return Div(Class("flex flex-wrap gap-1.5"), gomponents.Group(badges))
}
