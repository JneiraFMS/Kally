// Package kbfile définit le format de fichier Markdown d'une KB
// (`# ID`, front matter `tags:`, sections `## Erreur` / `## Notes` /
// `## Résolution`). C'est le codec unique partagé par le dépôt de
// persistance serveur (adapters/outbound/persistence) et les cas d'usage
// de la bibliothèque locale (File System Access) : aucune réimplémentation
// client ne peut dériver du format canonique.
package kbfile

import (
	"fmt"
	"regexp"
	"strings"

	domainkb "grz/internal/domain/kb"
)

// Serialize rend une KB dans le format fichier canonique.
func Serialize(item domainkb.KB) string {
	var sb strings.Builder
	if len(item.Tags) > 0 {
		sb.WriteString("---\ntags: [")
		sb.WriteString(strings.Join(item.Tags, ", "))
		sb.WriteString("]\n---\n\n")
	}
	fmt.Fprintf(&sb, "# %s\n\n## Erreur\n%s\n\n## Notes\n%s\n\n## Résolution\n%s\n",
		item.ID,
		escapeUserHeadings(item.Erreur),
		escapeUserHeadings(item.Notes),
		escapeUserHeadings(item.Resolution))
	return sb.String()
}

// Parse lit un contenu Markdown en KB. L'identifiant vient du nom de
// fichier (le `# ID` du corps n'est jamais une source de vérité) et une
// heure de modification éventuelle reste hors codec (stat du fichier).
func Parse(id, raw string) domainkb.KB {
	item := domainkb.KB{ID: id}
	body, tags := parseFrontMatter(raw)
	item.Tags = tags

	var section string
	var hasSection bool
	var hasLine bool
	var content strings.Builder
	flush := func() {
		if !hasSection {
			return
		}
		// Les headings utilisateur (échappées à l'écriture) reviennent
		// telles quelles ; une section inconnue reste ignorée, mais elle
		// ne peut plus être introduite par du contenu utilisateur.
		text := unescapeUserHeadings(strings.Trim(content.String(), "\n"))
		switch section {
		case "Erreur":
			item.Erreur = text
		case "Notes":
			item.Notes = text
		case "Résolution":
			item.Resolution = text
		}
		content.Reset()
		hasLine = false
	}
	for {
		line, rest, more := strings.Cut(body, "\n")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			flush()
			section = strings.TrimPrefix(trimmed, "## ")
			hasSection = true
		} else if hasSection {
			if hasLine {
				content.WriteByte('\n')
			}
			content.WriteString(line)
			hasLine = true
		}
		if !more {
			break
		}
		body = rest
	}
	flush()
	return item
}

// escapeUserHeadings neutralise le contenu utilisateur qui commence une
// ligne par « ## » : ce préfixe délimite les sections propres du fichier,
// une heading saisie dans un champ éclatait donc la section propriétaire et
// faisait perdre tout ce qui la suivait. Un anté-backslash invalide la
// frontière (et constitue un échappement de heading markdown valide) ;
// unescapeUserHeadings restitue le contenu à la lecture. Une ligne isolée
// `\## …` littéralement saisie par l'utilisateur s'affiche comme « ## … » :
// nuance acceptée, aucun contenu n'est perdu dans le round-trip.
func escapeUserHeadings(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			lines[i] = "\\" + line
		}
	}
	return strings.Join(lines, "\n")
}

// unescapeUserHeadings restitue les headings utilisateur échappées à
// l'écriture (voir escapeUserHeadings).
func unescapeUserHeadings(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "\\## ") {
			lines[i] = line[1:]
		}
	}
	return strings.Join(lines, "\n")
}

var frontMatterTagsPattern = regexp.MustCompile(`^tags:\s*\[([^\]]*)\]`)

func parseFrontMatter(raw string) (body string, tags []string) {
	if !strings.HasPrefix(raw, "---\n") {
		return raw, nil
	}
	rest := raw[4:]
	for {
		line, next, more := strings.Cut(rest, "\n")
		if !more {
			return raw, nil
		}
		if line == "---" {
			return next, tags
		}
		if match := frontMatterTagsPattern.FindStringSubmatch(strings.TrimSpace(line)); match != nil {
			for _, tag := range strings.Split(match[1], ",") {
				if tag = strings.TrimSpace(tag); tag != "" {
					tags = append(tags, tag)
				}
			}
		}
		rest = next
	}
}
