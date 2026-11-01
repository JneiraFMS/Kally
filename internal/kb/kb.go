package kb

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
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

// idPattern autorise uniquement lettres, chiffres, tirets et underscores :
// l'ID sert de nom de fichier, il ne doit donc pas contenir de séparateurs.
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
}

// NewKB valide les champs (ID et tags) puis retourne une KB prête à l'emploi.
func NewKB(id, erreur, notes, resolution string, rawTags ...string) (KB, error) {
	id = strings.TrimSpace(id)

	if id == "" || !idPattern.MatchString(id) {
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

// toMarkdown sérialise la KB au format Markdown avec front matter (tags).
func (kb KB) toMarkdown() string {
	var sb strings.Builder

	if len(kb.Tags) > 0 {
		sb.WriteString("---\ntags: [")
		sb.WriteString(strings.Join(kb.Tags, ", "))
		sb.WriteString("]\n---\n\n")
	}

	fmt.Fprintf(&sb,
		"# %s\n\n## Erreur\n%s\n\n## Notes\n%s\n\n## Résolution\n%s\n",
		kb.ID, kb.Erreur, kb.Notes, kb.Resolution,
	)

	return sb.String()
}

// parseMarkdown reconstruit une KB depuis son contenu .md.
func parseMarkdown(id, raw string) KB {
	kb := KB{ID: id}

	body, tags := parseFrontMatter(raw)
	kb.Tags = tags

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

	for _, line := range strings.Split(body, "\n") {
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

// parseFrontMatter extrait les tags du bloc YAML en tête de fichier
// (format Obsidian : --- / tags: [a, b] / ---) et retourne le reste du corps.
func parseFrontMatter(raw string) (body string, tags []string) {
	if !strings.HasPrefix(raw, "---\n") {
		return raw, nil
	}

	var fmLines []string
	rest := raw[4:]
	for {
		idx := strings.Index(rest, "\n")
		if idx == -1 {
			return raw, nil // pas de "---" fermant : pas de front matter valide
		}
		line := rest[:idx]
		if line == "---" {
			rest = rest[idx+1:]
			break
		}
		fmLines = append(fmLines, line)
		rest = rest[idx+1:]
	}

	re := regexp.MustCompile(`^tags:\s*\[([^\]]*)\]`)
	for _, line := range fmLines {
		if m := re.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			parts := strings.Split(m[1], ",")
			for _, p := range parts {
				if p = strings.TrimSpace(p); p != "" {
					tags = append(tags, p)
				}
			}
		}
	}

	return rest, tags
}
