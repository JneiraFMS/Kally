// Package journal définit l'entité du journal quotidien : une note libre
// (markdown) identifiée par sa seule date, à la Logseq.
package journal

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var (
	// ErrInvalidDate est retourné quand la date n'est pas au format YYYY-MM-DD
	// (séparateurs, bornes hors calendrier, ou valeur vide).
	ErrInvalidDate = errors.New("journal: date invalide (format attendu YYYY-MM-DD)")

	// ErrEmptyContent est retourné quand le contenu de la note est vide.
	ErrEmptyContent = errors.New("journal: contenu vide")

	// ErrNotFound est retourné quand l'entrée n'existe pas dans le repository.
	ErrNotFound = errors.New("journal: introuvable")
)

// DateLayout est le format canonique de la date d'une entrée.
const DateLayout = "2006-01-02"

// Entry est l'entité du domaine : une note libre datée. La date sert
// d'identité (une note par jour) ; il n'existe ni ID ni tags : les liens
// entre notes de journal suivent l'ordre des jours.
type Entry struct {
	Date    string
	Content string
}

// ValidateDate vérifie la forme canonique de la date et retourne
// ErrInvalidDate (mentionnant la valeur) sinon. Une date invalide est aussi
// un garde-fou d'attaque pour les adapters qui l'infèrent en nom de fichier.
func ValidateDate(date string) error {
	date = strings.TrimSpace(date)
	if _, err := time.Parse(DateLayout, date); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidDate, strings.TrimSpace(date))
	}
	return nil
}

// New valide la date et le contenu puis retourne une Entry prête à l'emploi.
func New(date, content string) (Entry, error) {
	date = strings.TrimSpace(date)
	if err := ValidateDate(date); err != nil {
		return Entry{}, err
	}
	if strings.TrimSpace(content) == "" {
		return Entry{}, ErrEmptyContent
	}
	return Entry{Date: date, Content: content}, nil
}

// EntriesDesc retourne les entrées triées de la plus récente à la plus
// ancienne (les dates ISO se comparent lexicalement), sans muter la source.
func EntriesDesc(entries []Entry) []Entry {
	sorted := slices.Clone(entries)
	slices.SortStableFunc(sorted, func(a, b Entry) int {
		return strings.Compare(b.Date, a.Date)
	})
	return sorted
}

// NeighborDates retourne les dates entourant date dans le journal : prev est
// le jour précédent (plus ancien), next le jour suivant (plus récent). Une
// valeur est vide quand le voisin n'existe pas ; une date absente du
// journal n'a aucun voisin.
func NeighborDates(entries []Entry, date string) (prev, next string) {
	sorted := EntriesDesc(entries)
	for i, entry := range sorted {
		if entry.Date != date {
			continue
		}
		if i+1 < len(sorted) {
			prev = sorted[i+1].Date // plus ancien, juste en dessous
		}
		if i > 0 {
			next = sorted[i-1].Date // plus récent, juste au dessus
		}
	}
	return prev, next
}
