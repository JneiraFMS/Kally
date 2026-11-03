package pages

import (
	"strconv"
	"strings"

	"grz/internal/application/dto"
	domainjournal "grz/internal/domain/journal"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageJournal affiche la page journal (inspirée Logseq) : la note du jour —
// créable en un clic quand elle n'existe pas, ouvrable sinon — puis la liste
// des entrées de la plus récente à la plus ancienne.
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
			todayCard(page.Today, page.TodayExists),
			entriesCard(page.Entries),
		),
	}

	return layout("Journal", currentPath, content)
}

// todayCard : la note du jour — lien « Ouvrir » vers sa fiche si elle
// existe, sinon bouton principal vers le formulaire (date du jour
// préselectionnée côté serveur).
func todayCard(today string, todayExists bool) gomponents.Node {
	var action gomponents.Node
	if todayExists {
		action = A(Href("/journal/"+today),
			Class("px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 font-medium transition"),
			gomponents.Text("Ouvrir la note du jour"),
		)
	} else {
		action = A(Href("/journal/new"),
			Class("px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 font-medium transition"),
			gomponents.Text("Créer la note du jour"),
		)
	}
	titre := "Note du jour"
	if today != "" {
		titre = formatDateFR(today)
	}
	return Div(Class("bg-g-widget rounded-xl p-6 flex items-center justify-between gap-4"),
		H2(Class("text-sm font-semibold text-g-hi font-mono"), gomponents.Text(titre)),
		action,
	)
}

// entriesCard liste les entrées du journal, de la plus récente à la plus
// ancienne (préparées par le service), avec un extrait de leur markdown.
func entriesCard(entries []domainjournal.Entry) gomponents.Node {
	if len(entries) == 0 {
		return Div(Class("bg-g-widget rounded-xl p-8 text-center"),
			P(Class("text-g-text"),
				gomponents.Text("Pas encore d'entrée : la note du jour est le meilleur point de départ.")),
		)
	}

	var rows []gomponents.Node
	for _, entry := range entries {
		rows = append(rows,
			Div(Class("py-2 flex items-baseline gap-3"),
				A(Href("/journal/"+entry.Date),
					Class("font-mono text-g-accent-text font-semibold hover:underline"),
					gomponents.Text(entry.Date),
				),
				Span(Class("text-g-text text-xs truncate"),
					gomponents.Text(firstLine(entry.Content)),
				),
			),
		)
	}

	return Div(Class("bg-g-widget rounded-xl px-6 py-4"),
		H2(Class("text-sm font-semibold text-g-hi pb-2"), gomponents.Text("Entrées précédentes")),
		Div(Class("divide-y divide-g-sep"), gomponents.Group(rows)),
	)
}

// formatDateFR affiche une date canonique en clair (« 3 novembre 2026 ») ;
// toute valeur non canonique est rendue brute (jamais d'index hors bornes).
func formatDateFR(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}
	jour, err := strconv.Atoi(parts[2])
	if err != nil || jour < 1 || jour > 31 {
		return date
	}
	mois, err := strconv.Atoi(parts[1])
	if err != nil || mois < 1 || mois > len(moisFR) {
		return date
	}
	return strconv.Itoa(jour) + " " + moisFR[mois-1] + " " + parts[0]
}

var moisFR = []string{
	"janvier", "février", "mars", "avril", "mai", "juin",
	"juillet", "août", "septembre", "octobre", "novembre", "décembre",
}
