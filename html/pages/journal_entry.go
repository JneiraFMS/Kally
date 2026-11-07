package pages

import (
	"grz/internal/application/dto"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageJournalEntry affiche une note de journal comme une note libre datée
// (inspiré Logseq) : titre en clair, markdown rendu, navigation jour
// précédent / suivant, actions Modifier et Supprimer.
func PageJournalEntry(currentPath string, page dto.JournalEntryPage) gomponents.Node {
	date := page.Entry.Date

	content := []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 pt-5"),
			Div(Class("max-w-3xl flex justify-end gap-2"),
				secondaryAction("/journal/"+date+"/edit", "Modifier"),
				secondaryAction("/journal", "Retour"),
				deleteJournalButton(date),
			),
		),
		Div(Class("max-w-5xl mx-auto px-6 pb-6"),
			Div(Class("max-w-3xl bg-g-widget rounded-xl px-8 py-8 space-y-6"),
				H1(Class("text-xl font-semibold text-g-hi"), gomponents.Text(formatDateFR(date))),
				mdBody(page.Entry.Content, nil),
			),
		),
		journalNav(page),
		codeCopyScriptNode(),
	}

	return layout("Journal du "+formatDateFR(date), currentPath, content)
}

// journalNav rend la navigation inter-journaux du détail : le jour
// précédent (plus ancien) à gauche, le suivant (plus récent) à droite. Un
// voisin absent laisse un pourtour invisible pour garder l'alignement.
func journalNav(page dto.JournalEntryPage) gomponents.Node {
	var prev, next gomponents.Node
	if page.Prev != "" {
		prev = A(Href("/journal/"+page.Prev),
			Class("px-3 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
			gomponents.Text("‹ Jour précédent"),
		)
	} else {
		prev = Span(Class("px-3 py-1.5 text-xs opacity-0 select-none"), gomponents.Text("‹ Jour précédent"))
	}
	if page.Next != "" {
		next = A(Href("/journal/"+page.Next),
			Class("px-3 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
			gomponents.Text("Jour suivant ›"),
		)
	} else {
		next = Span(Class("px-3 py-1.5 text-xs opacity-0 select-none"), gomponents.Text("Jour suivant ›"))
	}
	return Div(Class("max-w-5xl mx-auto px-6 pb-12"),
		Div(Class("max-w-3xl flex justify-between"), prev, next),
	)
}

// deleteJournalButton est le bouton de suppression d'une entrée journal :
// formulaire natif POST /journal/delete (fonctionne sans JavaScript),
// confirmation par le dialogue natif du navigateur (onsubmit + confirm) —
// un seul dialogue, avec comme sans htmx.
func deleteJournalButton(date string) gomponents.Node {
	return Form(
		Action("/journal/delete"),
		Method("post"),
		gomponents.Attr("hx-post", "/journal/delete"),
		gomponents.Attr("hx-disabled-elt", "find button"),
		gomponents.Attr("onsubmit", "return window.confirm('Supprimer la note du "+formatDateFR(date)+" ?')"),
		Class("mx-2 mb-1"),
		Input(Type("hidden"), Name("date"), Value(date)),
		Button(Type("submit"),
			Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-negative-text hover:bg-g-hover hover:brightness-125 text-xs transition-colors"),
			gomponents.Text("Supprimer"),
		),
	)
}
