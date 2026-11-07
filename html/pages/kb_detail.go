package pages

import (
	kb "grz/internal/domain/kb"
	"net/url"
	"strings"

	"grz/internal/application/dto"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBDetail affiche une KB comme une note : titre mono, tags en pills
// sous le titre, références liées et KB liées (tags + liens explicites). La
// forme suit le contenu réel : une section sans matière (Erreur/Notes/
// Résolution) ne s'affiche pas, de sorte qu'une KB créée en mode libre garde
// sa forme d'origine jusqu'à ce qu'elle soit complétée.
func PageKBDetail(currentPath string, detail dto.KBDetail) gomponents.Node {
	k := detail.KB

	var sections []gomponents.Node
	addSection := func(titre, contenu string, body func(string) gomponents.Node) {
		if strings.TrimSpace(contenu) != "" {
			sections = append(sections, noteSection(titre, body(contenu)))
		}
	}
	addSection("Erreur", k.Erreur, monoBody)
	addSection("Notes", k.Notes, func(texte string) gomponents.Node {
		return mdBody(texte, detail.KnownIDs)
	})
	addSection("Résolution", k.Resolution, func(texte string) gomponents.Node {
		return mdBody(texte, detail.KnownIDs)
	})
	sections = append(sections,
		backlinksSection(detail.Backlinks),
		relatedSection(detail.Related, detail.Neighbors),
	)

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
// POST /kb/delete (fonctionne sans JavaScript), confirmation par le
// dialogue natif du navigateur (onsubmit + confirm) — un seul dialogue,
// avec comme sans htmx (hx-confirm en doublon en demandait deux quand
// htmx intercepte le submit). Style danger : texte rouge.
func deleteButton(id string) gomponents.Node {
	return Form(
		Action("/kb/delete"),
		Method("post"),
		gomponents.Attr("hx-post", "/kb/delete"),
		gomponents.Attr("hx-disabled-elt", "find button"),
		// La valeur est échappée par gomponents (attribut sûr).
		gomponents.Attr("onsubmit", "return window.confirm('Supprimer définitivement "+id+" ?')"),
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
// « md » pour les styles). knownIDs résout les liens wiki [[ID]] : nil rend
// tous les liens existants, une liste active la détection des liens morts.
// Vide → « — vide — ».
func mdBody(texte string, knownIDs []string) gomponents.Node {
	if strings.TrimSpace(texte) == "" {
		return I(Class("text-g-text"), gomponents.Text("— vide —"))
	}
	return mdToNodeWithKnown(texte, knownIDs)
}

// noteSection rend une section : petit titre gras + corps. Les sections sont
// séparées par l'espace seul (pas de bordure) pour rester minimalistes.
func noteSection(titre string, body gomponents.Node) gomponents.Node {
	return Div(Class("space-y-1.5"),
		H2(Class("text-sm font-semibold text-g-hi"), gomponents.Text(titre)),
		body,
	)
}

// backlinksSection liste les KB qui citent celle-ci avec [[...]] (relation
// orientée : la source n'a pas besoin de citer en retour).
func backlinksSection(backlinks []kb.KB) gomponents.Node {
	if len(backlinks) == 0 {
		return noteSection("Références liées",
			I(Class("text-g-text"), gomponents.Text("Aucune référence liée")),
		)
	}

	var choices []gomponents.Node
	for _, k := range backlinks {
		choices = append(choices, kbEntry(k))
	}

	return noteSection("Références liées",
		Div(Class("divide-y divide-g-sep"), gomponents.Group(choices)),
	)
}

// kbEntry rend une KB dans une liste : lien interne, article ServiceNow et
// extrait de l'erreur.
func kbEntry(k kb.KB) gomponents.Node {
	return kbEntryWithBadge(k, false)
}

// kbEntryWithBadge rend une KB dans une liste ; linked ajoute un badge
// « lien » signalant une KB citée avec [[...]] sans partager de tag.
func kbEntryWithBadge(k kb.KB, linked bool) gomponents.Node {
	var badges []gomponents.Node
	if linked {
		badges = append(badges, Span(
			Class("inline-block px-2 py-0.5 rounded-full bg-g-block text-g-text text-xs"),
			gomponents.Attr("title", "Lien explicite [[…]]"),
			gomponents.Text("lien"),
		))
	}
	return Div(Class("py-1.5"),
		Div(Class("flex items-baseline gap-2"),
			// Navigation interne bidirectionnelle, conservée.
			A(Href("/kb/detail?id="+k.ID),
				Class("font-mono text-g-accent-text font-semibold hover:underline"),
				gomponents.Text(k.ID),
			),
			// Article ServiceNow de la KB liée, en lien secondaire.
			serviceNowLink(k.ID, "ServiceNow ↗",
				"text-xs text-g-text hover:text-g-hi hover:underline"),
			gomponents.Group(badges),
		),
		Div(Class("text-g-text text-xs"), gomponents.Text(firstLine(k.Erreur))),
	)
}

// relatedSection liste les KB liées à la KB principale : celles qui partagent
// un tag et celles qu'elle cite explicitement avec [[...]] (badge « lien »).
// Les liens entrants vivent dans « Références liées », jamais dupliqués ici.
func relatedSection(related, neighbors []kb.KB) gomponents.Node {
	if len(neighbors) == 0 {
		return noteSection("KB liées",
			I(Class("text-g-text"), gomponents.Text("Aucune KB liée")),
		)
	}

	relatedIDs := make(map[string]struct{}, len(related))
	for _, k := range related {
		relatedIDs[k.ID] = struct{}{}
	}

	var choices []gomponents.Node
	for _, k := range neighbors {
		_, sharesTag := relatedIDs[k.ID]
		choices = append(choices, kbEntryWithBadge(k, !sharesTag))
	}

	return noteSection("KB liées",
		Div(Class("divide-y divide-g-sep"), gomponents.Group(choices)),
	)
}
