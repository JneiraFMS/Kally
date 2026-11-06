package pages

import "testing"

// TestModuleSAPAndTCodesExports : les dérivations exposées aux adapters
// (bibliothèque locale) doivent retomber exactement sur le comportement
// des cartes serveur — sinon les deux types de cartes divergent visuellement.
func TestModuleSAPAndTCodesExports(t *testing.T) {
	if got := ModuleSAPFor([]string{"mm", "cc_flow"}); got != "SAP MM" {
		t.Errorf("ModuleSAPFor() = %q, want SAP MM", got)
	}
	if got := ModuleSAPFor([]string{"sap", "cc_flow", "po_creation_change"}); got != "" {
		t.Errorf("ModuleSAPFor(tags non module) = %q, want vide", got)
	}

	codes := TCodesFor("passer /nME22N pour ajuster", "VA01 aussi touché")
	if len(codes) != 2 || codes[0] != "ME22N" || codes[1] != "VA01" {
		t.Errorf("TCodesFor() = %v, want [ME22N VA01]", codes)
	}
	if codes := TCodesFor("", ""); codes != nil {
		t.Errorf("TCodesFor(vide) = %v, want nil", codes)
	}
}
