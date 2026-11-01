package kb

import (
	"errors"
	"slices"
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

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr error
	}{
		{name: "vide", input: "   ", want: []string{}},
		{
			name:  "séparés par virgules et espaces",
			input: " incident ,  calcul kpi, bug ",
			want:  []string{"incident", "calcul", "kpi", "bug"},
		},
		{
			name:  "majuscules normalisées",
			input: "Incident-KPI",
			want:  []string{"incident-kpi"},
		},
		{
			name:  "doublons dédupliqués",
			input: "a a b",
			want:  []string{"a", "b"},
		},
		{
			name:    "caractères interdits",
			input:   "a;b",
			wantErr: ErrInvalidTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTags(tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NormalizeTags() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeTags() inattendu: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("NormalizeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKBWithTags(t *testing.T) {
	kb, err := NewKB("KB-0001", "e", "n", "r", " Incident, kpi ")
	if err != nil {
		t.Fatalf("NewKB() inattendu: %v", err)
	}
	want := []string{"incident", "kpi"}
	if !slices.Equal(kb.Tags, want) {
		t.Errorf("NewKB() Tags = %v, want %v", kb.Tags, want)
	}

	_, err = NewKB("KB-0002", "e", "n", "r", "tag;invalide")
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("NewKB() tag invalide: error = %v, want ErrInvalidTag", err)
	}
}
