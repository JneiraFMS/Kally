package kb

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalidID est retourné quand l'identifiant d'une KB est vide ou
	// contient des caractères interdits.
	ErrInvalidID = errors.New("kb: identifiant invalide (vide ou caractères interdits)")

	// ErrNotFound est retourné quand une KB n'existe pas dans le repository.
	ErrNotFound = errors.New("kb: introuvable")
)

// idPattern autorise uniquement lettres, chiffres, tirets et underscores :
// l'ID sert de nom de fichier, il ne doit donc pas contenir de séparateurs.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// MVP : Value Object entité du domaine.
type KB struct {
	ID         string
	Erreur     string
	Notes      string
	Resolution string
}

// NewKB valide les champs puis retourne une KB prête à l'emploi.
func NewKB(id, erreur, notes, resolution string) (KB, error) {
	id = strings.TrimSpace(id)

	if id == "" || !idPattern.MatchString(id) {
		return KB{}, ErrInvalidID
	}

	return KB{
		ID:         id,
		Erreur:     erreur,
		Notes:      notes,
		Resolution: resolution,
	}, nil
}

// toMarkdown sérialise la KB au format Markdown.
func (kb KB) toMarkdown() string {
	return fmt.Sprintf(
		"# %s\n\n## Erreur\n%s\n\n## Notes\n%s\n\n## Résolution\n%s\n",
		kb.ID, kb.Erreur, kb.Notes, kb.Resolution,
	)
}

// parseMarkdown reconstruit une KB depuis son contenu .md.
func parseMarkdown(id, raw string) KB {
	kb := KB{ID: id}

	var section *string
	var lines []string

	flush := func() {
		if section == nil {
			return
		}
		text := strings.Trim(strings.Join(lines, "\n"), "\n")
		switch *section {
		case "Erreur":
			kb.Erreur = text
		case "Notes":
			kb.Notes = text
		case "Résolution":
			kb.Resolution = text
		}
	}

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			flush()
			name := strings.TrimPrefix(trimmed, "## ")
			section = &name
			lines = nil
			continue
		}

		lines = append(lines, line)
	}
	flush()

	return kb
}
