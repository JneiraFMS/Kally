package pages

import (
	kb "grz/internal/domain/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// tableHeaders : colonnes de la vue tableau haute densité.
// L'ID est cliquable (vers le détail) ; la date est en tabular-nums.
var tableHeaders = []string{"ID", "Module", "Titre", "Tag", "T-codes", "Date"}

// kbTable : la vue tableau haute densité — en-têtes discrets (fond bloc
// quasi-neutre), lignes h-9 avec hover, ID mono cliquable, module dérivé,
// titre + 1re ligne d'erreur tronqués, tag principal, premier T-code,
// date tabular-nums. Le conteneur porte data-fuzzy-view="table" : le
// moteur fuzzy ne compte chaque vue qu'une fois (voir fuzzyScript), et la
// bascule (viewScript) la cache par défaut.
func kbTable(kbs []kb.KB) gomponents.Node {
	rows := make([]gomponents.Node, 0, len(kbs))
	for _, k := range kbs {
		rows = append(rows, kbTableRow(k))
	}

	var headCells []gomponents.Node
	for _, h := range tableHeaders {
		headCells = append(headCells, Th(
			Scope("col"),
			// Entête discret : petites capitales espacées, fond bloc porté
			// par la ligne (kbTable).
			Class("px-4 py-2.5 text-left text-xs text-g-text font-medium uppercase tracking-wider"),
			gomponents.Text(h),
		))
	}

	return Div(
		ID("kb-table"),
		// Cachée par la bascule (home_search.go) ; dé-shown, le conteneur
		// est plus large que la grille (max-w-7xl) : 6 colonnes denses
		// respirent, le titre tronqué profite de l'écran.
		Class("hidden max-w-7xl mx-auto px-6 pb-8"),
		gomponents.Attr("data-fuzzy-view", "table"),
		// Surface carte : cadre arrondi bordé, en-têtes sur fond bloc,
		// cellules qui respirent (px-4) — densité conservée (h-9, text-xs).
		Div(Class("overflow-hidden rounded-xl border border-g-border bg-g-widget"),
			Table(Class("w-full text-sm border-collapse text-left"),
				THead(
					Tr(
						Class("bg-g-block border-b border-g-border"),
						gomponents.Group(headCells),
					),
				),
				gomponents.Group(rows),
			),
		),
	)
}

// kbTableRow : une ligne dense. Clic possible sur l'ID (le lien étendu
// overlay n'existe pas en tableau : la navigation passe par l'ID mono).
func kbTableRow(k kb.KB) gomponents.Node {
	return Tr(
		Class("h-9 hover:bg-g-hover transition-colors border-b border-g-border"),
		Data("search", searchText(k)),
		Td(Class("px-4 font-mono text-xs"),
			A(Href("/kb/detail?id="+k.ID),
				Class("text-g-accent-text hover:underline"),
				// Le tiroir fonctionne aussi depuis la vue tableau
				gomponents.Attr("data-kb-drawer", k.ID),
				gomponents.Text(k.ID))),
		Td(Class("px-4 text-xs text-g-text"), gomponents.Text(moduleSAPLabel(k))),
		Td(Class("px-4 text-g-text"),
			Div(Class("max-w-2xl truncate"), gomponents.Text(excerpt(k.Erreur, 90)))),
		Td(Class("px-4 font-mono text-[11px] text-g-text"), gomponents.Text(firstTagOrNull(k))),
		Td(Class("px-4 font-mono text-[11px] text-g-text"), gomponents.Text(firstTCodeOrNull(k))),
		Td(Class("px-4 text-xs text-g-text"),
			gomponents.If(!k.Modified.IsZero(),
				Span(Class("tabular-nums"), gomponents.Text(k.Modified.Format("2006-01-02"))))),
	)
}

// moduleSAPLabel : libellé module, tiret absent si non dérivable.
func moduleSAPLabel(k kb.KB) string {
	if m := moduleSAP(k.Tags); m != "" {
		return m
	}
	return "—"
}

// firstTagOrNull : premier tag, « — » sans tag.
func firstTagOrNull(k kb.KB) string {
	if len(k.Tags) > 0 {
		return k.Tags[0]
	}
	return "—"
}

// firstTCodeOrNull : premier T-code dérivé, « — » si aucun.
func firstTCodeOrNull(k kb.KB) string {
	if codes := kBTcodes(k); len(codes) > 0 {
		return codes[0]
	}
	return "—"
}
