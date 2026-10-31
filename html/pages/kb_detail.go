package pages

import (
	"grz/internal/kb"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBDetail affiche une KB comme une note Obsidian (vue lecture) :
// titre H1, tags cliquables, sections H2, et les KB liées par tags.
func PageKBDetail(currentPath string, k kb.KB, related []kb.KB) gomponents.Node {
	sections := []gomponents.Node{noteSection("Erreur", bodyOrEmpty(k.Erreur))}

	if len(k.Tags) > 0 {
		sections = append(sections, noteSection("Tags", tagBadges(k.Tags)))
	}

	sections = append(sections,
		noteSection("Notes", bodyOrEmpty(k.Notes)),
		noteSection("Résolution", bodyOrEmpty(k.Resolution)),
		relatedSection(related),
	)

	return layout("KB "+k.ID, currentPath, []gomponents.Node{
		// Barre d'actions discrète (style Obsidian : interfaces en fond uni)
		Div(Class("max-w-3xl mx-auto px-4 pt-6 flex justify-end gap-2"),
			A(Href("/kb/edit?id="+k.ID),
				Class("px-3 py-1.5 rounded-md border border-gray-200 bg-white text-sm text-gray-600 hover:bg-gray-100"),
				gomponents.Text("Modifier"),
			),
			A(Href("/kb/graph"),
				Class("px-3 py-1.5 rounded-md border border-gray-200 bg-white text-sm text-gray-600 hover:bg-gray-100"),
				gomponents.Text("Graphe"),
			),
			A(Href("/kb"),
				Class("px-3 py-1.5 rounded-md border border-gray-200 bg-white text-sm text-gray-600 hover:bg-gray-100"),
				gomponents.Text("← Liste"),
			),
		),

		// La "note"
		Note(k.ID, sections...),
	})
}

// Note est le conteneur papier d'une note Obsidian.
func Note(titre string, sections ...gomponents.Node) gomponents.Node {
	return Div(Class("max-w-3xl mx-auto px-4 pb-10"),
		Div(Class("bg-white ring-1 ring-gray-200 rounded-lg shadow-sm px-10 py-8 space-y-6"),
			H1(Class("text-3xl font-bold text-gray-900"), gomponents.Text(titre)),
			gomponents.Group(sections),
		),
	)
}

// bodyOrEmpty rend le corps d'une section texte ; vide → "— vide —".
func bodyOrEmpty(texte string) gomponents.Node {
	if strings.TrimSpace(texte) == "" {
		return I(Class("text-gray-400"), gomponents.Text("— vide —"))
	}
	return Div(Class("text-[15px] leading-relaxed text-gray-800 whitespace-pre-wrap"), gomponents.Text(texte))
}

// noteSection rend une section "à la Obsidian" : H2 suivi du corps.
func noteSection(titre string, body gomponents.Node) gomponents.Node {
	return Div(Class("space-y-1"),
		H2(Class("text-xl font-bold text-gray-900 border-b border-gray-200 pb-1"),
			gomponents.Text(titre),
		),
		body,
	)
}

// relatedSection liste les KB partageant au moins un tag (liens bidirectionnels).
func relatedSection(related []kb.KB) gomponents.Node {
	if len(related) == 0 {
		return noteSection("KB liées",
			I(Class("text-gray-400"), gomponents.Text("Aucune KB liée par tag")),
		)
	}

	var choices []gomponents.Node
	for _, k := range related {
		choices = append(choices,
			Div(Class("py-1"),
				A(Href("/kb/detail?id="+k.ID),
					Class("font-mono text-blue-700 font-semibold hover:underline"),
					gomponents.Text(k.ID),
				),
				Span(Class("text-gray-500 text-sm mx-2"), gomponents.Text("·")),
				Span(Class("text-gray-700 text-sm"), gomponents.Text(firstLine(k.Erreur))),
			),
		)
	}

	return noteSection("KB liées",
		Div(Class("text-[15px] divide-y divide-gray-100"), gomponents.Group(choices)),
	)
}
