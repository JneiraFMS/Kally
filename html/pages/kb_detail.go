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
		noteSection("Erreur", monoBody(k.Erreur)),
		noteSection("Notes", mdBody(k.Notes)),
		noteSection("Résolution", mdBody(k.Resolution)),
		relatedSection(related),
	}

	return layout("KB "+k.ID, currentPath, []gomponents.Node{
		// Barre d'actions discrète (boutons secondaires sans bordure,
		// suppression en rouge avec confirmation).
		Div(Class("max-w-3xl mx-auto px-4 pt-5 flex justify-end gap-2"),
			secondaryAction("/kb/edit?id="+k.ID, "Modifier"),
			secondaryAction("/kb/graph", "Graphe"),
			secondaryAction("/", "Retour"),
			deleteButton(k.ID),
		),

		Note(k.ID, k.Tags, sections...),
	})
}

// deleteButton est le bouton de suppression d'une KB : formulaire natif
// POST /kb/delete (fonctionne sans JavaScript), confirmation fournie par
// htmx (hx-confirm) quand il est actif. Style danger : texte rouge.
func deleteButton(id string) gomponents.Node {
	return Form(
		Action("/kb/delete"),
		Method("post"),
		gomponents.Attr("hx-post", "/kb/delete"),
		gomponents.Attr("hx-confirm", "Supprimer définitivement "+id+" ?"),
		gomponents.Attr("hx-disabled-elt", "find button"),
		Class("mx-2 mb-1"),
		Input(Type("hidden"), Name("id"), Value(id)),
		Button(Type("submit"),
			Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-negative hover:bg-g-hover hover:brightness-125 text-xs transition-colors"),
			gomponents.Text("Supprimer"),
		),
	)
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

// monoBody rend le corps de la section Erreur : un <pre> monospace
// (traces de pile alignées, frappées ligne à ligne), texte échappé par
// gomponents.Text. Vide → « — vide — ».
func monoBody(texte string) gomponents.Node {
	if strings.TrimSpace(texte) == "" {
		return I(Class("text-g-base"), gomponents.Text("— vide —"))
	}
	return Pre(Class("font-mono text-xs leading-relaxed whitespace-pre-wrap break-words text-g-text"),
		gomponents.Text(texte))
}

// mdBody rend une section Notes/Résolution comme du markdown sûr (classe
// « md » pour les styles). Vide → « — vide — ».
func mdBody(texte string) gomponents.Node {
	if strings.TrimSpace(texte) == "" {
		return I(Class("text-g-base"), gomponents.Text("— vide —"))
	}
	return mdToNode(texte)
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
			I(Class("text-g-base"), gomponents.Text("Aucune KB liée par tag")),
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
