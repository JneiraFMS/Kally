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
// des entrées de la plus récente à la plus ancienne, filtrable par la
// recherche fuzzy partagée (100 % client).
func PageJournal(currentPath string, page dto.JournalPage) gomponents.Node {
	content := []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6 flex items-center justify-between"),
			H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Journal")),
			A(Href("/"),
				Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
				gomponents.Text("Retour"),
			),
		),
		Div(Class("max-w-5xl mx-auto px-6 pb-8 space-y-4"), gomponents.Group(append(
			[]gomponents.Node{todayCard(page.Today, page.TodayExists)},
			entriesSection(page.Entries)...,
		))),
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

// entriesSection rassemble la partie « entrées » : barre de recherche fuzzy
// partagée + liste indexée (data-search par ligne) + état « aucun résultat »
// + script client quand des entrées existent ; sans entrée, un état vide
// clair (ni barre ni script : rien à filtrer).
func entriesSection(entries []domainjournal.Entry) []gomponents.Node {
	if len(entries) == 0 {
		return []gomponents.Node{emptyEntriesCard()}
	}

	var rows []gomponents.Node
	for _, entry := range entries {
		rows = append(rows, entryRow(entry))
	}

	return []gomponents.Node{
		searchField("journal-search", "journal-search-count", "journal-no-results",
			"Rechercher une entrée… (touche / pour filtrer)",
			"Aucune entrée correspondant à", false),
		Div(Class("bg-g-widget rounded-xl px-6 py-4"),
			H2(Class("text-sm font-semibold text-g-hi pb-2"), gomponents.Text("Entrées précédentes")),
			Div(Class("divide-y divide-g-sep"), gomponents.Group(rows)),
		),
		noResultsBlock("journal-no-results", ""),
		searchScriptNode(),
	}
}

// emptyEntriesCard reprend l'ancien état vide de la liste d'entrées.
func emptyEntriesCard() gomponents.Node {
	return Div(Class("bg-g-widget rounded-xl p-8 text-center"),
		P(Class("text-g-text"),
			gomponents.Text("Pas encore d'entrée : la note du jour est le meilleur point de départ.")),
	)
}

// entryRow liste une entrée : date cliquable + extrait du markdown ; la clé
// data-search alimente le filtre fuzzy client (dates + début du contenu).
func entryRow(entry domainjournal.Entry) gomponents.Node {
	return Div(Data("search", journalSearchKey(entry)),
		Class("py-2 flex items-baseline gap-3"),
		A(Href("/journal/"+entry.Date),
			Class("font-mono text-g-accent-text font-semibold hover:underline"),
			gomponents.Text(entry.Date),
		),
		Span(Class("text-g-text text-xs truncate"),
			gomponents.Text(firstLine(entry.Content)),
		),
	)
}

// journalSearchKey construit la clé de filtre d'une entrée : date brute,
// date en clair et les 3 premières lignes non vides du markdown, tout en
// minuscules (le scoreur compare sur un texte déjà abaissé, comme l'index
// des KB). Garder l'index court garde le DOM léger sur des notes longues.
func journalSearchKey(entry domainjournal.Entry) string {
	lines := []string{}
	for _, line := range strings.Split(entry.Content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) == 3 {
			break
		}
	}
	parts := append([]string{entry.Date, formatDateFR(entry.Date)}, lines...)
	return strings.TrimSpace(strings.ToLower(strings.Join(parts, " ")))
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
