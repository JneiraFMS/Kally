package pages

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// tagBadges rend les tags d'une KB en pills neutres cliquables vers le
// filtre /?tag=xxx. Vide pour une KB sans tag.
func tagBadges(tags []string) gomponents.Node {
	var badges []gomponents.Node
	for _, tag := range tags {
		badges = append(badges,
			A(Href("/?tag="+tag),
				Class("inline-block px-2.5 py-0.5 rounded-full bg-g-block text-g-base text-xs hover:text-g-hi hover:bg-g-hover transition-colors"),
				gomponents.Text(tag),
			),
		)
	}

	return Div(Class("flex flex-wrap gap-1.5"), gomponents.Group(badges))
}
