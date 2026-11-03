package pages

import (
	"net/url"

	"grz/internal/application/dto"
	domainkb "grz/internal/domain/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageJournal affiche la page journal (inspirée Logseq) : un bloc « note du
// jour » créable en un clic (le formulaire de création arrive prérempli avec
// l'ID journal-<date> et le tag journal), puis la liste des entrées du
// journal de la plus récente à la plus ancienne.
func PageJournal(currentPath string, page dto.JournalPage) gomponents.Node {
	content := []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6 flex items-center justify-between"),
			H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Journal")),
			A(Href("/"),
				Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
				gomponents.Text("Retour"),
			),
		),
		Div(Class("max-w-5xl mx-auto px-6 pb-8 space-y-4"),
			todayCard(page.TodayEntry, page.TodayID),
			entriesCard(page.Entries),
		),
	}

	return layout("Journal", currentPath, content)
}

// todayCard : la note du jour — lien « Ouvrir » vers la fiche si elle
// existe, sinon bouton principal vers la création préremplie.
func todayCard(entry *domainkb.KB, todayID string) gomponents.Node {
	var action gomponents.Node
	var resume gomponents.Node
	if entry != nil {
		if extrait := firstLine(entry.Erreur); extrait != "" {
			resume = P(Class("text-g-text text-sm"),
				gomponents.Text(extrait))
		}
		action = A(Href("/kb/detail?id="+url.QueryEscape(todayID)),
			Class("px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 font-medium transition"),
			gomponents.Text("Ouvrir la note du jour"),
		)
	} else {
		action = A(Href("/kb/new?id="+url.QueryEscape(todayID)+"&tags=journal"),
			Class("px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 font-medium transition"),
			gomponents.Text("Créer la note du jour"),
		)
	}

	return Div(Class("bg-g-widget rounded-xl p-6 flex items-center justify-between gap-4"),
		Div(Class("space-y-1"),
			H2(Class("text-sm font-semibold text-g-hi font-mono"), gomponents.Text(todayID)),
			resume,
		),
		action,
	)
}

// entriesCard liste les entrées du journal, de la plus récente à la plus
// ancienne (ID journal-<date> trié en descendant).
func entriesCard(entries []domainkb.KB) gomponents.Node {
	if len(entries) == 0 {
		return Div(Class("bg-g-widget rounded-xl p-8 text-center"),
			P(Class("text-g-text"),
				gomponents.Text("Pas encore d'entrée : la note du jour est le meilleur point de départ.")),
		)
	}

	var rows []gomponents.Node
	for _, e := range entries {
		rows = append(rows,
			Div(Class("py-2 flex items-baseline gap-3"),
				A(Href("/kb/detail?id="+e.ID),
					Class("font-mono text-g-accent-text font-semibold hover:underline"),
					gomponents.Text(e.ID),
				),
				Span(Class("text-g-text text-xs truncate"),
					gomponents.Text(firstLine(e.Erreur)),
				),
			),
		)
	}

	return Div(Class("bg-g-widget rounded-xl px-6 py-4"),
		H2(Class("text-sm font-semibold text-g-hi pb-2"), gomponents.Text("Entrées précédentes")),
		Div(Class("divide-y divide-g-sep"), gomponents.Group(rows)),
	)
}
