package dto

import domainkb "grz/internal/domain/kb"

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

// JournalPage is the daily-journal page data (inspired by Logseq):
// entries newest first, plus the state of today's note.
type JournalPage struct {
	// Today est la date du jour, ex. "2026-11-03".
	Today string
	// TodayID est l'identifiant attendu de la note du jour (journal-<date>).
	TodayID string
	// TodayExists signale que la note du jour a déjà été créée.
	TodayExists bool
	// TodayEntry est la note du jour, nil si elle n'existe pas encore.
	TodayEntry *domainkb.KB
	// Entries est la liste des entrées du journal, les plus récentes d'abord.
	Entries []domainkb.KB
}
