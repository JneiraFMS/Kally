package pages

import (
	"grz/internal/kb"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// firstLine retourne la première ligne non vide d'un texte (résumé court).
func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}

// searchText construit la clé de recherche fuzzy d'une KB : ID, première
// ligne de l'erreur et tags, tout en minuscules. Elle est embarquée dans
// l'attribut data-search de chaque carte pour que le filtrage se fasse
// 100 % côté client (aucun aller-retour réseau à la frappe).
func searchText(k kb.KB) string {
	return strings.ToLower(k.ID + " " + firstLine(k.Erreur) + " " + strings.Join(k.Tags, " "))
}

// PageAccueil affiche la page d'accueil : cartes des KB existantes,
// filtrées par tag si filtreTag est renseigné (chip "Filtré par" affiché).
func PageAccueil(currentPath string, kbs []kb.KB, filtreTag string) gomponents.Node {
	var cards []gomponents.Node
	for _, k := range kbs {
		cards = append(cards, kbCard(k))
	}

	header := []gomponents.Node{
		Div(Class("flex items-center justify-between"),
			H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Bases de connaissance")),
			A(Href("/kb/new"),
				Class("px-4 py-2 rounded-lg text-[#002b36] bg-g-accent hover:brightness-110 font-medium transition"),
				gomponents.Text("Créer une KB"),
			),
		),
	}

	// Tag de filtre actif, avec lien pour le retirer
	if filtreTag != "" {
		header = append(header,
			Div(Class("inline-flex items-center gap-2 px-3 py-1 rounded-full bg-g-block text-g-base text-xs"),
				gomponents.Text("Filtré par"),
				Span(Class("font-mono text-g-hi"), gomponents.Text(filtreTag)),
				A(Href("/"), Class("text-g-muted hover:text-g-hi"), gomponents.Text("×")),
			),
		)
	}

	// Barre de recherche fuzzy : rendue seulement s'il y a des cartes
	// (les états vides n'ont rien à filtrer). Le filtrage est client.
	hasKBs := len(kbs) > 0
	if hasKBs {
		header = append(header,
			Div(Class("relative"),
				Input(
					ID("kb-search"),
					Type("search"),
					AutoComplete("off"),
					Placeholder("Recherche fuzzy… (touche / pour filtre instantané)"),
					Class("w-full px-4 py-2.5 rounded-lg bg-g-block border border-g-border text-g-hi placeholder-g-subdue focus:border-g-accent outline-none transition"),
				),
				Span(ID("kb-search-count"),
					Class("absolute right-3 top-1/2 -translate-y-1/2 font-mono text-xs text-g-subdue pointer-events-none"),
					gomponents.Text(""),
				),
			),
		)
	}

	content := []gomponents.Node{
		Div(Class("p-6 max-w-5xl mx-auto space-y-4"), gomponents.Group(header)),
	}

	switch {
	case len(kbs) == 0 && filtreTag != "":
		// Aucune KB pour ce tag : invitation à voir toutes les KB
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-6 pb-8"),
				Div(Class("bg-g-widget rounded-xl p-10 text-center space-y-2"),
					P(Class("text-g-base"), gomponents.Text("Aucune KB avec le tag « "+filtreTag+" ».")),
					A(Href("/"), Class("text-g-base hover:text-g-hi hover:underline"), gomponents.Text("Voir toutes les KB")),
				),
			),
		)
	case len(kbs) == 0:
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-6 pb-8"),
				Div(Class("bg-g-widget rounded-xl p-10 text-center space-y-2"),
					P(Class("text-g-base"), gomponents.Text("Aucune KB enregistrée.")),
					A(Href("/kb/new"), Class("text-g-base hover:text-g-hi hover:underline"), gomponents.Text("Créer une première KB")),
				),
			),
		)
	default:
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-6 pb-8 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6"),
				gomponents.Group(cards),
			),
		)
	}

	// État « aucun résultat » togglé par le script, et script du scoreur
	// fuzzy injecté seulement quand il y a des cartes à filtrer.
	if hasKBs {
		content = append(content,
			Div(ID("kb-no-results"), Class("hidden max-w-5xl mx-auto px-6 pb-8"),
				Div(Class("bg-g-widget rounded-xl p-10 text-center text-g-base"), gomponents.Text("")),
			),
			searchScriptNode(),
		)
	}

	return layout("GRZ - Bases de connaissance", currentPath, content)
}

// kbCard est la carte d'une KB : ID en mono jaune, extrait de l'erreur,
// tags en pills. Sans bordure : l'élévation vient du fond widget, le survol
// remonte d'un ton. Technique "stretched link" : la carte est un <div> et un
// lien invisible (<a absolute inset-0>) couvre toute la carte, ce qui la rend
// cliquable vers le détail SANS imbriquer de <a> dans un <a> (HTML invalide).
// Les pills tags sont au-dessus de l'overlay (relative z-10) et restent
// cliquables vers le filtre /?tag=xxx.
func kbCard(k kb.KB) gomponents.Node {
	return Div(
		Class("relative bg-g-widget rounded-xl p-5 space-y-3 hover:bg-g-hover transition-colors"),
		// Clé de recherche fuzzy (évaluée côté client, voir home_search.go)
		Data("search", searchText(k)), // html.Data → data-search="..."
		// Lien invisible couvrant toute la carte
		A(
			Href("/kb/detail?id="+k.ID),
			Class("absolute inset-0 rounded-xl"),
		),
		Span(Class("inline-block px-2.5 py-0.5 rounded-full bg-g-block font-mono text-xs font-semibold text-g-accent"),
			gomponents.Text(k.ID),
		),
		P(Class("text-g-text text-sm line-clamp-3"),
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
