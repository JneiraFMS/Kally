package pages

import (
	kb "grz/internal/domain/kb"
	"net/url"
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
		relatedSection(k, related),
	}

	return layout("KB "+k.ID, currentPath, []gomponents.Node{
		// Barre d'actions discrète (boutons secondaires sans bordure,
		// suppression en rouge avec confirmation).
		Div(Class("max-w-5xl mx-auto px-6 pt-5"),
			Div(Class("max-w-3xl flex justify-end gap-2"),
				secondaryAction("/kb/edit?id="+k.ID, "Modifier"),
				secondaryAction("/kb/graph", "Graphe"),
				secondaryAction("/", "Retour"),
				deleteButton(k.ID),
			),
		),

		Note(k.ID, k.Tags, sections...),
		codeCopyScriptNode(),
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
			Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-negative-text hover:bg-g-hover hover:brightness-125 text-xs transition-colors"),
			gomponents.Text("Supprimer"),
		),
	)
}

// secondaryAction est un bouton d'action discret (bloc sans bordure).
func secondaryAction(href, label string) gomponents.Node {
	return A(Href(href),
		Class("mx-2 mb-1 px-3.5 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
		gomponents.Text(label),
	)
}

// serviceNowURL construit l'URL de l'article ServiceNow d'une KB.
func serviceNowURL(id string) string {
	return "https://airbus.service-now.com/kb_view.do?sysparm_article=" + url.QueryEscape(id)
}

// serviceNowLink rend un lien vers l'article ServiceNow d'une KB : ouverture
// dans un nouvel onglet, dont le libellé et le style sont fournis par
// l'appelant (titre de note vs entrée des KB liées).
func serviceNowLink(id, label, class string) gomponents.Node {
	return A(
		Href(serviceNowURL(id)),
		Target("_blank"),
		Rel("noopener noreferrer"),
		Class(class),
		gomponents.Text(label),
	)
}

// noteTitle : le titre de note est un H1 (sémantique). Pour un ID de KB,
// le texte est un lien vers l'article ServiceNow, héritant la couleur du h1.
func noteTitle(titre string) gomponents.Node {
	if strings.HasPrefix(titre, "KB") {
		return H1(Class("font-mono text-xl font-semibold text-g-hi"),
			serviceNowLink(titre, titre, "hover:underline"))
	}
	return H1(Class("font-mono text-xl font-semibold text-g-hi"), gomponents.Text(titre))
}

// Note est le conteneur d'une note : titre mono + tags en pills, puis les sections.
func Note(titre string, tags []string, sections ...gomponents.Node) gomponents.Node {
	var sousTitre []gomponents.Node
	if len(tags) > 0 {
		sousTitre = append(sousTitre, tagBadges(tags))
	}

	return Div(Class("max-w-5xl mx-auto px-6 pb-10"),
		Div(Class("max-w-3xl bg-g-widget rounded-xl px-8 py-8 space-y-6"),
			Div(Class("space-y-2"),
				noteTitle(titre),
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
		return I(Class("text-g-text"), gomponents.Text("— vide —"))
	}
	return Pre(Class("font-mono text-xs leading-relaxed whitespace-pre-wrap break-words text-g-hi bg-g-code border border-g-border rounded-lg p-3"),
		gomponents.Text(texte))
}

// mdBody rend une section Notes/Résolution comme du markdown sûr (classe
// « md » pour les styles). Vide → « — vide — ».
func mdBody(texte string) gomponents.Node {
	if strings.TrimSpace(texte) == "" {
		return I(Class("text-g-text"), gomponents.Text("— vide —"))
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

// relatedSection liste les KB partageant au moins un tag avec la KB
// principale (liens bidirectionnels), avec en tête le mini-graphe local de
// la note (inspiré Logseq) quand il y a au moins un voisin.
func relatedSection(principal kb.KB, related []kb.KB) gomponents.Node {
	if len(related) == 0 {
		return noteSection("KB liées",
			I(Class("text-g-text"), gomponents.Text("Aucune KB liée par tag")),
		)
	}

	var choices []gomponents.Node
	for _, k := range related {
		choices = append(choices,
			Div(Class("py-1.5"),
				Div(Class("flex items-baseline gap-2"),
					// Navigation interne bidirectionnelle, conservée.
					A(Href("/kb/detail?id="+k.ID),
						Class("font-mono text-g-accent-text font-semibold hover:underline"),
						gomponents.Text(k.ID),
					),
					// Article ServiceNow de la KB liée, en lien secondaire.
					serviceNowLink(k.ID, "ServiceNow ↗",
						"text-xs text-g-text hover:text-g-hi hover:underline"),
				),
				Div(Class("text-g-text text-xs"), gomponents.Text(firstLine(k.Erreur))),
			),
		)
	}

	return noteSection("KB liées",
		Div(Class("space-y-4"),
			localGraph(principal, related),
			Div(Class("divide-y divide-g-sep"), gomponents.Group(choices)),
		),
	)
}
