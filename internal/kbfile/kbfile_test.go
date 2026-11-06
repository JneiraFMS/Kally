package kbfile

import (
	"strings"
	"testing"

	domainkb "grz/internal/domain/kb"
)

// TestSerializeWithTags vérifie le rendu canonique d'une KB avec front
// matter tags : le format fichier partagé par le dépôt serveur et la
// bibliothèque locale (File System Access).
func TestSerializeWithTags(t *testing.T) {
	item := domainkb.KB{ID: "KB-0001", Erreur: "boom", Tags: []string{"mm", "achat"}}
	want := "---\ntags: [mm, achat]\n---\n\n"
	want += "# KB-0001\n\n## Erreur\nboom\n\n## Notes\n\n\n## Résolution\n\n"
	if got := Serialize(item); got != want {
		t.Errorf("Serialize =\n%q\nwant\n%q", got, want)
	}
}

// TestParseRoundTripPreservesContent : une KB écrite puis relue retrouve
// exactement son contenu, y compris un contenu utilisateur qui commence
// une ligne par « ## » (échappé à l'écriture).
func TestParseRoundTripPreservesContent(t *testing.T) {
	notes := "Première ligne.\n\n## Pas une section du fichier, une heading utilisateur\n\nSuite.\n[[KB-0002]]"
	item, err := domainkb.NewKB("KB-0001", "trace", notes, "résolution", "mm")
	if err != nil {
		t.Fatal(err)
	}
	back := Parse("KB-0001", Serialize(item))
	if back.ID != "KB-0001" {
		t.Errorf("ID = %q, want KB-0001", back.ID)
	}
	if back.Erreur != item.Erreur {
		t.Errorf("Erreur := %q, want %q", back.Erreur, item.Erreur)
	}
	if back.Notes != notes {
		t.Errorf("Notes := %q, want %q", back.Notes, notes)
	}
	if back.Resolution != "résolution" {
		t.Errorf("Resolution = %q, want %q", back.Resolution, "résolution")
	}
	if strings.Join(back.Tags, ",") != "mm" {
		t.Errorf("Tags = %v, want [mm]", back.Tags)
	}
}

// TestParseSections : une section inconnue est ignorée, les sections
// connues vont au bon champ, la première ligne `# ID` est ignorée au
// profit de l'identifiant du fichier.
func TestParseSections(t *testing.T) {
	raw := "---\ntags: [a, b]\n---\n\n# NOT-KB\n\n## Erreur\ntrace\n\n## Notes\nnote\n\n## Inconnu\ntexte perdu\n\n## Résolution\nfix\n"
	item := Parse("KB-0001", raw)
	if item.ID != "KB-0001" {
		t.Errorf("ID = %q, want KB-0001", item.ID)
	}
	if item.Erreur != "trace" || item.Notes != "note" || item.Resolution != "fix" {
		t.Errorf("sections = %q / %q / %q", item.Erreur, item.Notes, item.Resolution)
	}
	if item.Tags[0] != "a" || item.Tags[1] != "b" {
		t.Errorf("Tags = %v, want [a b]", item.Tags)
	}
}

// TestParseWithoutFrontMatter : un fichier sans front matter parse sans tags.
func TestParseWithoutFrontMatter(t *testing.T) {
	item := Parse("KB-0001", "# KB-0001\n\n## Notes\ncontenu\n")
	if len(item.Tags) != 0 {
		t.Errorf("Tags = %v, want vide", item.Tags)
	}
	if item.Notes != "contenu" {
		t.Errorf("Notes = %q", item.Notes)
	}
}

// TestParseEscapedUserHeadingsAreRestored : une ligne `\## …` échappée
// revient telle quelle (round-trip asymétrique assumé, voir AGENT.md).
func TestParseEscapedUserHeadingsAreRestored(t *testing.T) {
	item := Parse("KB-0001", "## Notes\n\\## heading utilisateur\nsuite\n")
	want := "## heading utilisateur\nsuite"
	if item.Notes != want {
		t.Errorf("Notes = %q, want %q", item.Notes, want)
	}
}
