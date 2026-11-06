package pages

import (
	kb "grz/internal/domain/kb"
	"net/url"
	"slices"
	"sort"
	"strconv"
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

// PageAccueil affiche la page d'accueil : barre d'outils (titre, bouton
// secondaire), recherche command-palette, filtres de tags segmentés puis
// cartes des KB existantes (filtrées sur les tags fournis, AND à la
// Logseq : tous les tags requis).
func PageAccueil(currentPath string, kbs []kb.KB, filtreTags ...string) gomponents.Node {
	var cards []gomponents.Node
	for _, k := range kbs {
		cards = append(cards, kbCard(k))
	}
	// Normalisation défensive (production déjà normalisée par l'adapter) :
	// des valeurs vides/interdites doivent rester sémantiquement « sans filtre ».
	filtreTags = kb.NormalizeTagFilter(filtreTags)
	hasFiltre := len(filtreTags) > 0

	// Barre d'outils : titre de section + CTA secondaire bordé (mode
	// libre). Le CTA primaire « Créer une KB » vit dans la navbar.
	header := []gomponents.Node{
		Div(Class("flex items-start justify-between gap-4 flex-wrap"),
			Div(Class("space-y-0.5"),
				H1(Class("text-base font-semibold tracking-tight text-g-hi"),
					gomponents.Text("Bases de connaissances")),
				P(Class("text-xs text-g-text"),
					gomponents.Text("Fiches d'incidents SAP/ERP : symptômes, cause racine et résolution.")),
			),
			Div(Class("flex items-center gap-2"),
				secondaryAction("/kb/new/quick", "Nouvelle KB libre"),
			),
		),
	}

	// Recherche fuzzy et filtres segmentés : rendus seulement s'il y a des
	// cartes (les états vides n'ont rien à filtrer). Le filtrage est client.
	hasKBs := len(kbs) > 0
	if hasKBs {
		header = append(header,
			searchField("kb-search", "kb-search-count", "kb-no-results",
				"Recherche",
				"Aucune KB correspondant à", true),
			tagSegmentRow(kbs, filtreTags),
		)
	}

	content := []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6 space-y-4"), gomponents.Group(header)),
	}

	switch {
	case len(kbs) == 0 && hasFiltre:
		// Aucune KB pour ces tags : invitation à voir toutes les KB
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-6 pb-8"),
				Div(Class("bg-g-widget rounded-xl p-10 text-center space-y-2"),
					P(Class("text-g-text"), gomponents.Text("Aucune KB avec le(s) tag(s) « "+strings.Join(filtreTags, ", ")+" ».")),
					A(Href("/"), Class("text-g-text hover:text-g-hi hover:underline"), gomponents.Text("Voir toutes les KB")),
				),
			),
		)
	case len(kbs) == 0:
		content = append(content,
			// état vide togglé par grz-local : des articles locaux font
			// disparaître « Aucune KB enregistrée ».
			Div(ID("kb-empty"), Class("max-w-5xl mx-auto px-6 pb-8"),
				Div(Class("bg-g-widget rounded-xl p-10 text-center space-y-2"),
					P(Class("text-g-text"), gomponents.Text("Aucune KB enregistrée.")),
					A(Href("/kb/new"), Class("text-g-text hover:text-g-hi hover:underline"), gomponents.Text("Créer une première KB")),
				),
			),
		)
	default:
		content = append(content,
			// Bascule de vue (cartes/tableau), conteneurs distincts du
			// moteur fuzzy (comptage par vue visible), grille puis tableau.
			// Le SCRIPT de bascule est injecté plus bas, après la grille :
			// exécuté au parsing, il ne la verrait pas sinon.
			Div(Class("max-w-5xl mx-auto px-6 pb-8 flex justify-end gap-1.5"),
				viewToggleButtons(),
			),
			kbTable(kbs),
		)
	}

	// Grille des cartes : toujours rendue (même vide) — les articles
	// locaux de la bibliothèque FSA s'invitent à côté des KB partagées.
	// Un filtre actif suspend l'injection (les tags locaux ne sont pas
	// filtrables, le résultat resterait trompeur).
	content = append(content,
		Div(ID("kb-cards"),
			Class("max-w-5xl mx-auto px-6 pb-8 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5"),
			gomponents.Attr("data-fuzzy-view", "grid"),
			gomponents.If(hasFiltre, gomponents.Attr("data-tag-filtered", "1")),
			gomponents.Group(cards),
		),
	)

	// État « aucun résultat » togglé par le script, recherche fuzzy et
	// bascule de vue — injectés APRÈS la grille, qu'ils doivent trouver
	// dans le DOM au moment de s'exécuter.
	if hasKBs {
		content = append(content,
			noResultsBlock("kb-no-results", "max-w-5xl mx-auto px-6 pb-8"),
			searchScriptNode(),
			viewToggleScriptNode(),
		)
	}

	// Barre des articles locaux (grz-local) : connexion et état, en pied
	// de page — discrète, cachée par défaut (révélée seulement sur
	// navigateur compatible). Les cartes locales, elles, s'invitent dans
	// la grille partagée ci-dessus.
	// Shell du tiroir AVANT la barre : le script grz-local s'exécute au
	// parsing et capture les éléments du tiroir — le shell doit exister
	// avant lui (les fiches locales s'y ouvrent).
	content = append(content, drawerShell(), PageLocalBar())

	return layout("Kally - Bases de connaissances", currentPath, content)
}

// kbCard est la carte d'une KB, canon « console » : en-tête mono non
// encadré `ID · Module SAP · date` (séparateur point médian, date en
// tabular-nums à droite), titre = 1re ligne de l'erreur (line-clamp-2),
// extrait = lignes suivantes (line-clamp-2), pied : tag principal mono
// cliquable, premier T-code dérivé et action « Consulter → ». Surface
// g-widget liseré fin, le survol pousse le liseré vers l'accent.
// Technique "stretched link" : la carte est un <div> et un lien invisible
// (<a absolute inset-0>) couvre toute la carte — les éléments cliquables
// restent au-dessus (relative z-10).
func kbCard(k kb.KB) gomponents.Node {
	return Div(
		Class("relative bg-g-widget rounded-xl p-5 border border-g-border hover:border-g-cyan transition-colors"),
		// Clé de recherche fuzzy (évaluée côté client, voir home_search.go)
		Data("search", searchText(k)), // html.Data → data-search="..."
		// Lien invisible couvrant toute la carte : navigation native
		// (sans JS) et déclencheur du tiroir (data-kb-drawer, câblé par
		// le script grz-drawer)
		A(
			Href("/kb/detail?id="+k.ID),
			Class("absolute inset-0 rounded-xl"),
			gomponents.Attr("data-kb-drawer", k.ID),
		),
		cardHeader(k),
		cardTitleAndExcerpt(k),
		cardFooter(k),
	)
}

// cardHeader : en-tête zéro-pilule `ID · SAP MM · date` — l'ID en mono
// accentué, le module s'il est dérivable, la date en tabular-nums au bord
// droit. Rien n'est inventé : pas de module, pas de date, pas d'affichage.
func cardHeader(k kb.KB) gomponents.Node {
	left := []gomponents.Node{
		Span(Class("font-mono text-xs font-semibold text-g-accent-text"), gomponents.Text(k.ID)),
	}
	if module := moduleSAP(k.Tags); module != "" {
		left = append(left,
			Span(Class("text-g-subdue"), gomponents.Text("·")),
			Span(Class("text-xs text-g-text"), gomponents.Text(module)),
		)
	}

	nodes := []gomponents.Node{
		Div(Class("relative z-10 flex flex-wrap items-baseline gap-x-2 gap-y-0.5"),
			gomponents.Group(left),
		),
	}
	if !k.Modified.IsZero() {
		// La ligne de date vit dans la rangée d'en-tête, alignée à droite
		// par l'ordre du flux : elle complète la liste des méta.
		nodes[0] = Div(Class("relative z-10 flex flex-wrap items-baseline gap-x-2 gap-y-0.5 justify-between"),
			gomponents.Group(left),
			Span(Class("tabular-nums font-mono text-[11px] text-g-text"),
				gomponents.Text(k.Modified.Format("2006-01-02"))),
		)
	}
	return gomponents.Group(nodes)
}

// cardTitleAndExcerpt : titre = première ligne de l'erreur (semibold,
// line-clamp-2), extrait = lignes suivantes (text-xs, line-clamp-2,
// omis quand l'erreur tient sur une ligne).
func cardTitleAndExcerpt(k kb.KB) gomponents.Node {
	nodes := []gomponents.Node{}
	if titre := firstLine(k.Erreur); titre != "" {
		nodes = append(nodes,
			P(Class("relative z-0 text-sm font-semibold text-g-hi line-clamp-2 mt-2"),
				gomponents.Text(titre)))
	}
	if rest := afterFirstLine(k.Erreur); rest != "" {
		nodes = append(nodes,
			P(Class("relative z-0 text-xs text-g-text line-clamp-2 mt-1"),
				gomponents.Text(excerpt(rest, 140))))
	}
	return gomponents.Group(nodes)
}

// cardFooter : tag principal (premier de la liste) en chip mono discret
// cliquable, premier T-code dérivé, action « Consulter → » poussée à
// droite. Le pied est rendu même sans tag (l'action seule suffit).
func cardFooter(k kb.KB) gomponents.Node {
	items := []gomponents.Node{}
	if len(k.Tags) > 0 {
		items = append(items, cardTagChip(k.Tags[0]))
	}
	if codes := kBTcodes(k); len(codes) > 0 {
		items = append(items,
			Span(Class("font-mono text-[11px] text-g-text"), gomponents.Text(codes[0])),
		)
	}
	items = append(items,
		Span(Class("ml-auto text-xs text-g-text"), gomponents.Text("Consulter →")),
	)
	return Div(Class("relative z-10 flex flex-wrap items-center gap-2 mt-3"),
		gomponents.Group(items))
}

// viewToggleButtons : deux boutons icône (grille/tableau) portant
// data-view-target, l'état réparti en aria-pressed (script viewToggle).
func viewToggleButtons() gomponents.Node {
	button := func(target, label string, pressed bool, icon gomponents.Node) gomponents.Node {
		cls := "p-1.5 rounded-md border border-g-border bg-g-widget text-g-text hover:text-g-hi hover:border-g-cyan transition-colors"
		if pressed {
			cls += " bg-g-block text-g-hi"
		}
		return Button(
			Type("button"),
			gomponents.Attr("data-view-target", target),
			gomponents.Attr("aria-pressed", boolToStr(pressed)),
			gomponents.Attr("aria-label", label),
			Class(cls),
			icon,
		)
	}
	// L'état initial est rendu côté serveur compatible par défaut
	// (grilles actives) ; le script l'ajuste si le client a une vue en
	// mémoire (localStorage grz-view).
	pressedDefault := true
	return Div(Class("flex items-center gap-1"),
		button("grid", "Vue cartes", pressedDefault, gridIconSVG()),
		button("table", "Vue tableau", !pressedDefault, tableIconSVG()),
	)
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// gridIconSVG : icône grille (quatre carrés).
func gridIconSVG() gomponents.Node {
	return gomponents.El("svg",
		gomponents.Attr("viewBox", "0 0 24 24"),
		gomponents.Attr("width", "16"),
		gomponents.Attr("height", "16"),
		gomponents.Attr("aria-hidden", "true"),
		gomponents.Attr("fill", "none"),
		gomponents.El("path",
			gomponents.Attr("d", "M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z"),
			gomponents.Attr("stroke", "currentColor"), gomponents.Attr("stroke-width", "1.5"),
			gomponents.Attr("stroke-linecap", "round"), gomponents.Attr("stroke-linejoin", "round"),
		),
	)
}

// tableIconSVG : icône tableau (tailles-lignes).
func tableIconSVG() gomponents.Node {
	return gomponents.El("svg",
		gomponents.Attr("viewBox", "0 0 24 24"),
		gomponents.Attr("width", "16"),
		gomponents.Attr("height", "16"),
		gomponents.Attr("aria-hidden", "true"),
		gomponents.Attr("fill", "none"),
		gomponents.El("path",
			gomponents.Attr("d", "M4 5h16M4 10h16M4 15h16M4 19h16"),
			gomponents.Attr("stroke", "currentColor"), gomponents.Attr("stroke-width", "1.5"),
			gomponents.Attr("stroke-linecap", "round"),
		),
	)
}

// cardTagChip : chip mono discret pour un tag cliquable (filtre).
func cardTagChip(tag string) gomponents.Node {
	return A(Href("/?tag="+url.QueryEscape(tag)),
		Class("px-2 py-0.5 rounded bg-g-block font-mono text-[11px] text-g-text hover:text-g-hi hover:bg-g-hover transition-colors"),
		gomponents.Text(tag))
}

// afterFirstLine retourne les lignes suivant la première ligne non vide
// (l'extrait de carte : le titre est la première).
func afterFirstLine(text string) string {
	lines := strings.Split(text, "\n")
	seen := false
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			if seen {
				return strings.Join(lines[i:], "\n")
			}
			seen = true
		}
	}
	return ""
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

// tagSegmentRow agrège tous les tags des KB rendues, les trie, et les
// rend en boutons segmentés cliquables : un tag actif pointe vers l'URL
// SANS lui (retrait), un inactif vers l'URL avec lui (ajout). Zéro tag ⇒
// pas de rangée.
func tagSegmentRow(kbs []kb.KB, filtreTags []string) gomponents.Node {
	tags := allTags(kbs)
	if len(tags) == 0 {
		return nil
	}

	segments := make([]gomponents.Node, 0, len(tags))
	for _, tag := range tags {
		active := slices.Contains(filtreTags, tag)
		if active {
			segments = append(segments, tagSegment(filtreURLSans(filtreTags, tag), tag, true))
		} else {
			segments = append(segments, tagSegment(filtreURLAvec(filtreTags, tag), tag, false))
		}
	}

	return Div(Class("flex flex-wrap items-center gap-1.5"), gomponents.Group(segments))
}

// tagSegment est un bouton de filtre segmenté : mono discret en repos,
// plein accent quand actif (aria-pressed porte l'état pour l'AT).
func tagSegment(href, tag string, active bool) gomponents.Node {
	cls := "px-2.5 py-1 rounded-md border font-mono text-[11px] whitespace-nowrap transition-colors"
	if active {
		cls += " bg-g-accent text-g-ink border-transparent hover:brightness-110"
	} else {
		cls += " bg-g-widget text-g-text border-g-border hover:text-g-hi hover:border-g-cyan"
	}
	return A(Href(href), Class(cls), gomponents.Attr("aria-pressed", strconv.FormatBool(active)), gomponents.Text(tag))
}

// allTags retourne les tags distincts des KB, triés (ordre stable).
func allTags(kbs []kb.KB) []string {
	seen := map[string]bool{}
	var tags []string
	for _, k := range kbs {
		for _, tag := range k.Tags {
			if !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	sort.Strings(tags)
	return tags
}

// filtreURLAvec construit l'URL du filtre avec le tag ajouté.
func filtreURLAvec(tags []string, ajoute string) string {
	all := append(slices.Clone(tags), ajoute)
	return "/?" + url.Values{"tag": all}.Encode()
}

// filtreURLSans construit l'URL du filtre sans le tag retiré : sans tag
// restant, retour à l'accueil non filtré.
func filtreURLSans(tags []string, retire string) string {
	rest := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag != retire {
			rest = append(rest, tag)
		}
	}
	if len(rest) == 0 {
		return "/"
	}
	return "/?" + url.Values{"tag": rest}.Encode()
}
