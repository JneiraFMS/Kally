package dto

import (
	domainjournal "grz/internal/domain/journal"
	domainkb "grz/internal/domain/kb"
)

// KBInput is the application input used to create or update a knowledge base.
type KBInput struct {
	ID         string
	Erreur     string
	Notes      string
	Resolution string
	Tags       string
}

// KBDetail is the detail page data prepared by the application layer.
type KBDetail struct {
	KB      domainkb.KB
	Related []domainkb.KB
}

// JournalInput is the application input for a journal entry: a date
// (YYYY-MM-DD) plus free markdown content. No ID, no tags: the date is the
// identity of the note.
type JournalInput struct {
	Date    string
	Content string
}

// JournalPage is the daily-journal page data (inspired by Logseq):
// entries newest first, plus whether today's note exists yet.
type JournalPage struct {
	// Today est la date du jour, ex. "2026-11-03".
	Today string
	// TodayExists signale que la note du jour a déjà été créée.
	TodayExists bool
	// Entries est la liste des entrées du journal, les plus récentes d'abord.
	Entries []domainjournal.Entry
}

// JournalEntryPage is a single journal entry page data with day navigation:
// Prev est le jour précédent (plus ancien), Next le jour suivant (plus
// récent) ; une valeur vide signale l'absence de voisin.
type JournalEntryPage struct {
	Entry domainjournal.Entry
	Prev  string
	Next  string
}
