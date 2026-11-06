package persistence

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"grz/internal/application/ports/outbound"
	domainkb "grz/internal/domain/kb"
)

// MDRepository stores each knowledge base in a Markdown file.
type MDRepository struct {
	dir string
}

// NewMDRepository creates a repository rooted at dir.
func NewMDRepository(dir string) *MDRepository {
	return &MDRepository{dir: dir}
}

func (r *MDRepository) path(id string) string {
	return filepath.Join(r.dir, id+".md")
}

// Save creates or replaces the Markdown file for item.
func (r *MDRepository) Save(ctx context.Context, item domainkb.KB) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("kb md: %w", err)
	}
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("kb md: création du répertoire: %w", err)
	}
	if err := os.WriteFile(r.path(item.ID), []byte(toMarkdown(item)), 0o644); err != nil {
		return fmt.Errorf("kb md: écriture de %s: %w", item.ID, err)
	}
	return nil
}

// Exists reports whether the file for id exists.
func (r *MDRepository) Exists(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("kb md: %w", err)
	}
	if err := domainkb.ValidateID(id); err != nil {
		return false, domainkb.ErrInvalidID
	}
	_, err := os.Stat(r.path(id))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("kb md: stat %s: %w", id, err)
	}
	return true, nil
}

// Delete removes the file for id, or returns ErrNotFound when it is absent.
func (r *MDRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("kb md: %w", err)
	}
	if err := domainkb.ValidateID(id); err != nil {
		return domainkb.ErrInvalidID
	}
	if err := os.Remove(r.path(id)); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("kb md: %w: %s", domainkb.ErrNotFound, id)
	} else if err != nil {
		return fmt.Errorf("kb md: suppression %s: %w", id, err)
	}
	return nil
}

// FindAll reads Markdown files from the repository, sorted by ID.
func (r *MDRepository) FindAll(ctx context.Context) ([]domainkb.KB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("kb md: %w", err)
	}
	entries, err := os.ReadDir(r.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []domainkb.KB{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("kb md: lecture du répertoire: %w", err)
	}

	items := make([]domainkb.KB, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		// Les fichiers du journal (journal-<date>.md) cohabitent dans le
		// même répertoire mais ne sont pas des KB : ni accueil ni graphe.
		if strings.HasPrefix(id, journalFilePrefix) {
			continue
		}
		item, err := r.read(id)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	// os.ReadDir returns entries sorted by filename; stripping the common .md
	// suffix preserves the same order by ID without a second sort.
	return items, nil
}

// FindByID reads one knowledge base by ID.
func (r *MDRepository) FindByID(ctx context.Context, id string) (*domainkb.KB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("kb md: %w", err)
	}
	if err := domainkb.ValidateID(id); err != nil {
		return nil, domainkb.ErrInvalidID
	}
	return r.read(id)
}

func (r *MDRepository) read(id string) (*domainkb.KB, error) {
	data, err := os.ReadFile(r.path(id))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("kb md: %w: %s", domainkb.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("kb md: lecture %s: %w", id, err)
	}
	item := parseMarkdown(id, string(data))

	// Heure de dernière modification du fichier (mémoire de stockage, hors
	// format : rien n'est sérialisé dans le .md). Une stat qui échoue (ex:
	// fichier retiré entre ReadFile et Stat) laisse le champ à zéro — la
	// présentation sait masquer une date absente.
	if fi, statErr := os.Stat(r.path(id)); statErr == nil {
		item.Modified = fi.ModTime()
	}
	return &item, nil
}

func toMarkdown(item domainkb.KB) string {
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

func parseMarkdown(id, raw string) domainkb.KB {
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

var _ outbound.KBRepository = (*MDRepository)(nil)
