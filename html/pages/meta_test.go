package pages

import (
	"reflect"
	"testing"

	kb "grz/internal/domain/kb"
)

// TestModuleSAP : le module SAP d'une carte dérive d'un tag court (2-4
// lettres, normalisé en majuscules avec préfixe SAP) ; les tags de projet
// (cc_flow, underscore) et le tag générique « sap » sont ignorés.
func TestModuleSAP(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{"vide", nil, ""},
		{"un tag module", []string{"mm"}, "SAP MM"},
		{"tag sap ignoré", []string{"sap"}, ""},
		{"module après tag sap", []string{"sap", "mm"}, "SAP MM"},
		{"plusieurs modules : le premier", []string{"mm", "sd", "fi"}, "SAP MM"},
		{"tag projet ignoré", []string{"cc_flow"}, ""},
		{"tag projet seul", []string{"cc_flow", "quantite"}, ""},
		{"long tag ignoré", []string{"materiaux"}, ""},
		{"sd", []string{"sd"}, "SAP SD"},
		{"fi", []string{"fi"}, "SAP FI"},
		{"qm", []string{"qm"}, "SAP QM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := moduleSAP(tt.tags); got != tt.want {
				t.Errorf("moduleSAP(%v) = %q, want %q", tt.tags, got, tt.want)
			}
		})
	}
}

// TestExtractTCodes : les transactions SAP sont extraites des notes (pas
// de l'erreur, trace brute), dédupliquées dans l'ordre d'apparition,
// borne à 3.
func TestExtractTCodes(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"texte vide", "", nil},
		{"pas de code", "prix modifié manuellement dans le document", nil},
		{"code isolé", "via /nME22N", []string{"ME22N"}},
		{"code sans /n", "préférez MM02", []string{"MM02"}},
		{"code variante N", "créer par VA01N", []string{"VA01N"}},
		{"dédupliqué", "ME22N puis /nME22N encore", []string{"ME22N"}},
		{"ordre d'apparition", "ME21N et FB60", []string{"ME21N", "FB60"}},
		{"borne à 3", "ME21N ME22N MM02 FB60 MB1A", []string{"ME21N", "ME22N", "MM02"}},
		{"minuscule ignoré", "me22n dans une phrase", nil},
		{"collé à un mot", "surXME22N? pas un code", nil},
		{"notes et résolution combinées", "utiliser /nME02N ; sinon via MM03 révisons l'article", []string{"ME02N", "MM03"}},
		{"id KB cité non confondu", "voir KB12345", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTCodes(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTCodes(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

// TestKBTCodes : l'extraction s'appuie sur Notes + Résolution (l'Erreur
// est une trace brute jamais liée) ; les liens wiki [[ID]] ne gênent pas.
func TestKBTCodes(t *testing.T) {
	k, err := kb.NewKB("KB0001", "erreur brute sans code", "pousser /nME22N puis vérifier", "résolution", "mm")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	if got := kBTcodes(k); !reflect.DeepEqual(got, []string{"ME22N"}) {
		t.Errorf("kBTcodes(k) = %v, want [ME22N]", got)
	}
}
