package kb

import (
	"errors"
	"strings"
	"testing"
)

func TestNewKBValidation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
		wantID  string
	}{
		{name: "id vide", id: "   ", wantErr: ErrInvalidID},
		{name: "id avec slash (danger path traversal)", id: "../etc/passwd", wantErr: ErrInvalidID},
		{name: "id avec caractères spéciaux", id: "KB 001!", wantErr: ErrInvalidID},
		{
			name:   "id valide",
			id:     "KB-0001",
			wantID: "KB-0001",
		},
		{
			name:   "id avec espaces autour, trim",
			id:     "  KB-0002  ",
			wantID: "KB-0002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb, err := NewKB(tt.id, "erreur", "notes", "resolution")

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewKB() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewKB() inattendu: %v", err)
			}
			if kb.ID != tt.wantID {
				t.Errorf("NewKB() ID = %q, want %q", kb.ID, tt.wantID)
			}
			if kb.Erreur != "erreur" || kb.Notes != "notes" || kb.Resolution != "resolution" {
				t.Errorf("NewKB() champs incorrects: %+v", kb)
			}
		})
	}
}

func TestKBToMarkdown(t *testing.T) {
	kb, err := NewKB("KB-0001", "Division par zéro", "Regarder le diviseur", "Vérifier avant de diviser")
	if err != nil {
		t.Fatalf("NewKB() inattendu: %v", err)
	}

	md := kb.toMarkdown()

	for _, section := range []string{
		"# KB-0001",
		"## Erreur",
		"Division par zéro",
		"## Notes",
		"Regarder le diviseur",
		"## Résolution",
		"Vérifier avant de diviser",
	} {
		if !strings.Contains(md, section) {
			t.Errorf("toMarkdown() ne contient pas %q\nMarkdown:\n%s", section, md)
		}
	}
}
