package pages

import (
	"grz/internal/kb"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBDetail affiche une KB comme une note : titre mono, tags en pills
// sous le titre, sections séparées par l'espace seul, KB liées.
func PageKBDetail(currentPath string, k kb.KB, related []kb.KB) gomponents.Node {
	sections := []gomponents.Node{
		noteSection("Erreur", bodyOrEmpty(k.Erreur)),
		noteSection("Notes", bodyOrEmpty(k.Notes)),
		noteSection("Résolution", bodyOrEmpty(k.Resolution)),
		relatedSection(related),
	}

	return layout("KB "+k.ID, currentPath, []gomponents.Node{
		// Barre d'actions discrète (boutons secondaires sans bordure)
		Div(Class("max-w-3xl mx-auto px-4 pt-5 flex justify-end gap-2"),
			secondaryAction("/kb/edit?id="+k.ID, "Modifier"),
			secondaryAction("/kb/graph", "Graphe"),
			secondaryAction("/", "Retour"),
		),

		Note(k.ID, k.Tags, sections...),
	})
}

// secondaryAction est un bouton d'action discret (bloc sans bordure).
func secondaryAction(href, label string) gomponents.Node {
	return A(Href(href),
		Class("mx-2 mb-1 px-3.5 py-1.5 rounded-lg bg-g-block text-g-base hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
		gomponents.Text(label),
	)
}

// Note est le conteneur d'une note : titre mono + tags en pills, puis les sections.
func Note(titre string, tags []string, sections ...gomponents.Node) gomponents.Node {
	var sousTitre []gomponents.Node
	if len(tags) > 0 {
		sousTitre = append(sousTitre, tagBadges(tags))
	}

	return Div(Class("max-w-3xl mx-auto px-4 pb-10"),
		Div(Class("bg-g-widget rounded-xl px-8 py-8 space-y-6"),
			Div(Class("space-y-2"),
				H1(Class("font-mono text-xl font-semibold text-g-hi"), gomponents.Text(titre)),
				gomponents.Group(sousTitre),
			),
			gomponents.Group(sections),
		),
	)
}

// bodyOrEmpty rend le corps d'une section texte ; vide → "— vide —".
func bodyOrEmpty(texte string) gomponents.Node {
	if strings.TrimSpace(texte) == "" {
		return I(Class("text-g-subdue"), gomponents.Text("— vide —"))
	}
	return Div(Class("leading-relaxed text-g-text whitespace-pre-wrap"), gomponents.Text(texte))
}

// noteSection rend une section : petit titre gras + corps. Les sections sont
// séparées par l'espace seul (pas de bordure) pour rester minimalistes.
func noteSection(titre string, body gomponents.Node) gomponents.Node {
	return Div(Class("space-y-1.5"),
		H2(Class("text-sm font-semibold text-g-hi"), gomponents.Text(titre)),
		body,
	)
}

// relatedSection liste les KB partageant au moins un tag (liens bidirectionnels).
func relatedSection(related []kb.KB) gomponents.Node {
	if len(related) == 0 {
		return noteSection("KB liées",
			I(Class("text-g-subdue"), gomponents.Text("Aucune KB liée par tag")),
		)
	}

	var choices []gomponents.Node
	for _, k := range related {
		choices = append(choices,
			Div(Class("py-1.5"),
				A(Href("/kb/detail?id="+k.ID),
					Class("font-mono text-g-accent font-semibold hover:underline"),
					gomponents.Text(k.ID),
				),
				Div(Class("text-g-base text-xs"), gomponents.Text(firstLine(k.Erreur))),
			),
		)
	}

	return noteSection("KB liées",
		Div(Class("divide-y divide-g-sep"), gomponents.Group(choices)),
	)
}
