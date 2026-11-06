package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"grz/internal/application/dto"
	domainkb "grz/internal/domain/kb"
)

// La bibliothèque locale (File System Access) ne persiste rien sur le
// serveur : ces cas d'usage sont des transformations sans état autour du
// codec kbfile — sérialiser pour écrire le .md dans le dossier utilisateur,
// analyser un fichier lu depuis ce dossier, exposer les IDs connus du
// serveur pour distinguer un lien wiki vivant d'un lien mort.

func TestKBServiceSerializeLocalValidatesAndReturnsCanonicalMarkdown(t *testing.T) {
	service := NewKBService(newFakeRepository())
	input := dto.KBInput{ID: "KB-1", Erreur: "e", Notes: "n", Resolution: "r", Tags: "incident"}

	got, err := service.SerializeLocal(input)
	if err != nil {
		t.Fatalf("SerializeLocal() inattendu: %v", err)
	}
	if !strings.Contains(got, "---\ntags: [incident]\n---\n\n# KB-1\n") {
		t.Errorf("SerializeLocal() =\n%q", got)
	}

	if _, err := service.SerializeLocal(dto.KBInput{ID: "../bad"}); !errors.Is(err, domainkb.ErrInvalidID) {
		t.Errorf("SerializeLocal() ID invalide error = %v, want ErrInvalidID", err)
	}
	if _, err := service.SerializeLocal(dto.KBInput{ID: "KB-1", Tags: "no!"}); !errors.Is(err, domainkb.ErrInvalidTag) {
		t.Errorf("SerializeLocal() tag invalide error = %v, want ErrInvalidTag", err)
	}
}

func TestKBServiceParseLocalReturnsFieldsWithoutID(t *testing.T) {
	service := NewKBService(newFakeRepository())
	raw := "---\ntags: [mm]\n---\n\n# KB-1\n\n## Erreur\ntrace\n\n## Notes\nnote\n\n## Résolution\nfix\n"

	parsed, err := service.ParseLocal(raw)
	if err != nil {
		t.Fatalf("ParseLocal() inattendu: %v", err)
	}
	if parsed.Erreur != "trace" || parsed.Notes != "note" || parsed.Resolution != "fix" {
		t.Errorf("ParseLocal() = %+v", parsed)
	}
	if len(parsed.Tags) != 1 || parsed.Tags[0] != "mm" {
		t.Errorf("ParseLocal() tags = %v, want [mm]", parsed.Tags)
	}
}

func TestKBServiceLocalKnownIDsListsSharedKBsOnly(t *testing.T) {
	repo := newFakeRepository()
	repo.items["KB-1"] = domainKB(t, "KB-1", "incident")
	repo.items["KB-2"] = domainKB(t, "KB-2", "kpi")
	service := NewKBService(repo)

	ids, err := service.LocalKnownIDs(context.Background())
	if err != nil {
		t.Fatalf("LocalKnownIDs() inattendu: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("LocalKnownIDs() = %v", ids)
	}
}
