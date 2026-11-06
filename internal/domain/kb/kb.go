package kb

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
)

var (
	// ErrInvalidID est retourné quand l'identifiant d'une KB est vide ou
	// contient des caractères interdits.
	ErrInvalidID = errors.New("kb: identifiant invalide (vide ou caractères interdits)")

	// ErrInvalidTag est retourné quand un tag contient des caractères interdits.
	ErrInvalidTag = errors.New("kb: tag invalide (caractères interdits)")

	// ErrNotFound est retourné quand une KB n'existe pas dans le repository.
	ErrNotFound = errors.New("kb: introuvable")
)

// idPattern autorise uniquement lettres, chiffres, tirets et underscores.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// tagPattern restraint un tag validé (après normalisation lowercase).
var tagPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// KB est l'entité du domaine.
type KB struct {
	ID         string
	Erreur     string
	Notes      string
	Resolution string
	Tags       []string
	// Modified est l'heure de dernière modification du fichier source,
	// alimentée par l'adapter de persistance (stat du .md). Le domaine ne
	// la sérialise ni ne la consomme : zéro ⇔ non renseignée, et la
	// présentation n'affiche la date que si elle existe.
	Modified time.Time
}

// NewKB valide les champs (ID et tags) puis retourne une KB prête à l'emploi.
func NewKB(id, erreur, notes, resolution string, rawTags ...string) (KB, error) {
	id = strings.TrimSpace(id)

	if err := ValidateID(id); err != nil {
		return KB{}, ErrInvalidID
	}

	tags, err := NormalizeTags(rawTags...)
	if err != nil {
		return KB{}, err
	}

	return KB{
		ID:         id,
		Erreur:     erreur,
		Notes:      notes,
		Resolution: resolution,
		Tags:       tags,
	}, nil
}

// ValidateID reports whether id uses the allowed identifier characters.
func ValidateID(id string) error {
	if id == "" || !idPattern.MatchString(id) {
		return ErrInvalidID
	}
	return nil
}

// NormalizeTags découpe, trim, met en minuscules et dédoublonne les tags
// saisis (séparés par virgules et/ou espaces). Une erreur est retournée si un
// tag contient des caractères interdits.
func NormalizeTags(raw ...string) ([]string, error) {
	tags := []string{}

	for _, value := range raw {
		parts := strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || unicode.IsSpace(r)
		})
		for _, part := range parts {
			part = strings.ToLower(strings.TrimSpace(part))
			if part == "" {
				continue
			}
			if !tagPattern.MatchString(part) {
				return nil, fmt.Errorf("%w: %q", ErrInvalidTag, part)
			}
			if !slices.Contains(tags, part) {
				tags = append(tags, part)
			}
		}
	}

	return tags, nil
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
