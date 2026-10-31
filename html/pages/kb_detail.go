package pages

import (
	"grz/internal/kb"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBDetail affiche une KB comme une note Obsidian (vue lecture) :
// titre H1, sections H2, corps de texte lisible sur fond "papier".
func PageKBDetail(currentPath string, k kb.KB) gomponents.Node {
	return layout("KB "+k.ID, currentPath, []gomponents.Node{
		// Barre d'actions discrète (style Obsidian : interfaces en fond uni)
		Div(Class("max-w-3xl mx-auto px-4 pt-6 flex justify-end gap-2"),
			A(Href("/kb/edit?id="+k.ID),
				Class("px-3 py-1.5 rounded-md border border-gray-200 bg-white text-sm text-gray-600 hover:bg-gray-100"),
				gomponents.Text("Modifier"),
			),
			A(Href("/kb"),
				Class("px-3 py-1.5 rounded-md border border-gray-200 bg-white text-sm text-gray-600 hover:bg-gray-100"),
				gomponents.Text("← Liste"),
			),
		),

		// La "note"
		Note(k.ID, noteSection("Erreur", k.Erreur), noteSection("Notes", k.Notes), noteSection("Résolution", k.Resolution)),
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

// noteSection rend une section "à la Obsidian" : H2 suivi du texte, sauts de
// ligne conservés. Une section vide affiche une mention en italique gris.
func noteSection(titre string, texte string) gomponents.Node {
	body := gomponents.Text(texte)
	if strings.TrimSpace(texte) == "" {
		body = I(Class("text-gray-400"), gomponents.Text("— vide —"))
	}

	return Div(Class("space-y-1"),
		H2(Class("text-xl font-bold text-gray-900 border-b border-gray-200 pb-1"),
			gomponents.Text(titre),
		),
		Div(Class("text-[15px] leading-relaxed text-gray-800 whitespace-pre-wrap"), body),
	)
}
