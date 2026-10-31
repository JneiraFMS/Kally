package pages

import (
	"grz/internal/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageAccueil affiche la page d'accueil : cartes des KB existantes.
func PageAccueil(currentPath string, kbs []kb.KB) gomponents.Node {
	var cards []gomponents.Node
	for _, k := range kbs {
		cards = append(cards, kbCard(k))
	}

	content := []gomponents.Node{
		Div(Class("p-8 max-w-5xl mx-auto space-y-4"),
			Div(Class("flex items-center justify-between"),
				H1(Class("text-2xl font-bold"), gomponents.Text("Bases de connaissance")),
				A(Href("/kb/new"),
					Class("px-4 py-2 rounded text-white bg-blue-600 hover:bg-blue-700"),
					gomponents.Text("+ Nouvelle KB"),
				),
			),
		),
	}

	if len(cards) == 0 {
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-8 pb-8"),
				Div(Class("border border-dashed border-gray-300 rounded-lg p-10 text-center"),
					P(Class("text-gray-500"), gomponents.Text("Aucune KB enregistrée. Créez-en une !")),
					A(Href("/kb/new"),
						Class("mt-3 inline-block text-blue-600 hover:underline"),
						gomponents.Text("Créer une KB →"),
					),
				),
			),
		)
	} else {
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-8 pb-8 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"),
				gomponents.Group(cards),
			),
		)
	}

	return layout("GRZ - Bases de connaissance", currentPath, content)
}

// kbCard est la carte d'une KB : ID + extrait du message d'erreur.
// Technique "stretched link" : la carte est un <div> et un lien invisible
// (<a absolute inset-0>) couvre toute la carte, ce qui la rend cliquable
// vers le détail SANS imbriquer de <a> dans un <a> (HTML invalide).
// Les badges tags sont au-dessus de l'overlay (relative z-10) et restent
// cliquables vers le filtre /kb?tag=xxx.
func kbCard(k kb.KB) gomponents.Node {
	return Div(
		Class("relative border border-gray-200 rounded-lg shadow-sm hover:shadow-md hover:border-blue-300 transition p-4 space-y-2"),
		// Lien invisible couvrant toute la carte
		A(
			Href("/kb/detail?id="+k.ID),
			Class("absolute inset-0 rounded-lg"),
		),
		Div(Class("flex items-center justify-between"),
			Span(Class("px-2 py-0.5 rounded bg-blue-50 text-blue-700 text-sm font-mono font-semibold"),
				gomponents.Text(k.ID),
			),
			Span(Class("text-gray-400 text-sm"), gomponents.Text("→")),
		),
		P(Class("text-gray-700 text-sm line-clamp-3"),
			gomponents.Text(excerpt(k.Erreur, 120)),
		),
		Div(Class("relative z-10"), tagBadges(k.Tags)),
	)
}

// excerpt tronque un texte et ajoute "..." si besoin.
func excerpt(text string, max int) string {
	first := firstLine(text)
	r := []rune(first)
	if len(r) <= max {
		return first
	}
	return string(r[:max]) + "..."
}
