package kb

import (
	"slices"
	"strings"
)

// JournalPrefix est le préfixe d'ID des notes quotidiennes du journal
// (inspiré Logseq) : une entrée du jour a pour ID "journal-YYYY-MM-DD".
const JournalPrefix = "journal-"

// FilterByIDPrefix retourne les KB dont l'ID commence par prefix, triées de
// la plus récente (ID le plus grand) à la plus ancienne. Un préfixe vide ne
// retourne aucune KB (filtre explicite uniquement).
func FilterByIDPrefix(kbs []KB, prefix string) []KB {
	entries := []KB{}
	if prefix == "" {
		return entries
	}
	for _, k := range kbs {
		if strings.HasPrefix(k.ID, prefix) {
			entries = append(entries, k)
		}
	}
	slices.SortFunc(entries, func(a, b KB) int { return strings.Compare(b.ID, a.ID) })
	return entries
}

// NormalizeTagFilter nettoie un filtre de tags issu d'une requête HTTP :
// minuscules, espaces coupés, vides et doublons retirés. Un tag aux
// caractères interdits est ignoré (ne jamais échouer sur un filtre
// utilisateur ; il ne matchera simplement rien).
func NormalizeTagFilter(raw []string) []string {
	tags := []string{}
	for _, value := range raw {
		tag := strings.ToLower(strings.TrimSpace(value))
		if tag == "" || !tagPattern.MatchString(tag) {
			continue
		}
		if !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}
	return tags
}
