package kb

import (
	"slices"
	"testing"
)

func mustLinkKB(t *testing.T, id, erreur, notes, resolution string) KB {
	t.Helper()
	item, err := NewKB(id, erreur, notes, resolution)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return item
}

// TestLinkTargetsReadsNotesAndResolution : seuls Notes et Résolution sont
// markdown liables ; la trace Erreur reste du texte brut et le code est
// ignoré. L'ordre d'apparition est conservé, les doublons retirés.
func TestLinkTargetsReadsNotesAndResolution(t *testing.T) {
	k := mustLinkKB(t,
		"KB-A",
		"trace [[KB-ERR]] dans le message",
		"voir [[KB-B]] et `[[KB-CODE]]`",
		"[[KB-C]] puis [[KB-B]]",
	)

	got := LinkTargets(k)
	want := []string{"KB-B", "KB-C"}
	if !slices.Equal(got, want) {
		t.Errorf("LinkTargets() = %v, want %v", got, want)
	}
}

func TestLinkTargetsIgnoresInvalidIDs(t *testing.T) {
	k := mustLinkKB(t, "KB-A", "", "[[pas un id]] et [[KB_OK]]", "")

	got := LinkTargets(k)
	want := []string{"KB_OK"}
	if !slices.Equal(got, want) {
		t.Errorf("LinkTargets() = %v, want %v", got, want)
	}
}

// TestBacklinksAreDirected : A cite B, B ne cite personne ⇒ B affiche A en
// rétrolien, mais A n'affiche pas B. La relation n'est pas symétrique.
func TestBacklinksAreDirected(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-A", "", "[[KB-B]]", ""),
		mustLinkKB(t, "KB-B", "", "", ""),
	}

	if got, want := IDs(Backlinks(kbs, "KB-B")), []string{"KB-A"}; !slices.Equal(got, want) {
		t.Errorf("Backlinks(KB-B) = %v, want %v", got, want)
	}
	if got := Backlinks(kbs, "KB-A"); len(got) != 0 {
		t.Errorf("Backlinks(KB-A) = %v, want vide (B ne cite pas A)", IDs(got))
	}
}

func TestBacklinksExcludeSelfAndSort(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-B", "", "[[KB-A]]", ""),
		mustLinkKB(t, "KB-A", "", "[[KB-A]]", ""), // auto-référence ignorée
		mustLinkKB(t, "KB-C", "", "[[KB-A]]", ""),
		mustLinkKB(t, "KB-D", "", "", ""),
	}

	got := IDs(Backlinks(kbs, "KB-A"))
	want := []string{"KB-B", "KB-C"}
	if !slices.Equal(got, want) {
		t.Errorf("Backlinks(KB-A) = %v, want %v", got, want)
	}
}

func TestBacklinksIgnoreCode(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-A", "", "```\n[[KB-B]]\n```", ""),
		mustLinkKB(t, "KB-C", "", "`[[KB-B]]`", ""),
		mustLinkKB(t, "KB-B", "", "", ""),
	}

	if got := Backlinks(kbs, "KB-B"); len(got) != 0 {
		t.Errorf("Backlinks(KB-B) = %v, want vide (références en code ignorées)", IDs(got))
	}
}

func TestLinkEdgesExplicitAndBidirectional(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-B", "", "[[KB-A]]", ""),
		mustLinkKB(t, "KB-A", "", "[[KB-B]]", ""),
		mustLinkKB(t, "KB-C", "", "", ""),
	}

	edges := LinkEdges(kbs)
	want := []GraphEdge{{A: "KB-A", B: "KB-B", Linked: true}}
	if !slices.EqualFunc(edges, want, func(a, b GraphEdge) bool {
		return a.A == b.A && a.B == b.B && a.Linked == b.Linked
	}) {
		t.Errorf("LinkEdges() = %+v, want %+v (une arête dédupliquée)", edges, want)
	}
}

func TestLinkEdgesIgnoreMissingAndSelf(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-A", "", "[[KB-INCONNUE]] et [[KB-A]]", ""),
		mustLinkKB(t, "KB-B", "", "", ""),
	}

	if edges := LinkEdges(kbs); len(edges) != 0 {
		t.Errorf("LinkEdges() = %+v, want vide (cible absente et auto-lien ignorés)", edges)
	}
}

// TestLinkEdgesNormalizeOrder : une arête est non orientée (A<B) quel que
// soit le sens de la citation, et l'ordre de sortie est déterministe.
func TestLinkEdgesNormalizeOrder(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-Z", "", "[[KB-A]]", ""),
		mustLinkKB(t, "KB-A", "", "", ""),
	}

	got := LinkEdges(kbs)
	want := []GraphEdge{{A: "KB-A", B: "KB-Z", Linked: true}}
	if !slices.EqualFunc(got, want, func(a, b GraphEdge) bool {
		return a.A == b.A && a.B == b.B && a.Linked == b.Linked
	}) {
		t.Errorf("LinkEdges() = %+v, want %+v", got, want)
	}
}

// TestNeighborsUnionTagAndOutgoingLinks : le voisinage de la fiche réunit les
// KB liées par tag et les cibles sortantes [[...]], triées par ID. Les liens
// entrants restent dans Backlinks et une KB sans rapport reste exclue.
func TestNeighborsUnionTagAndOutgoingLinks(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-A", "", "[[KB-C]]", ""),
		mustLinkKB(t, "KB-B", "", "", ""),
		mustLinkKB(t, "KB-C", "", "", ""),
		mustLinkKB(t, "KB-D", "", "[[KB-A]]", ""),
		mustLinkKB(t, "KB-E", "", "", ""),
	}
	// KB-A et KB-B partagent un tag, KB-C est citée (sortant), KB-D cite
	// KB-A (entrant, donc hors « KB liées »).
	kbs[0].Tags = []string{"calcul"}
	kbs[1].Tags = []string{"calcul"}

	got := IDs(Neighbors(kbs, "KB-A"))
	want := []string{"KB-B", "KB-C"}
	if !slices.Equal(got, want) {
		t.Errorf("Neighbors(KB-A) = %v, want %v", got, want)
	}
}

func TestNeighborsIgnoreMissingTargets(t *testing.T) {
	kbs := []KB{
		mustLinkKB(t, "KB-A", "", "[[KB-INCONNUE]] et [[KB-A]]", ""),
	}

	if got := IDs(Neighbors(kbs, "KB-A")); len(got) != 0 {
		t.Errorf("Neighbors(KB-A) = %v, want vide (cible absente et auto-lien exclus)", got)
	}
}

func TestNeighborsUnknownKB(t *testing.T) {
	if got := Neighbors(nil, "KB-A"); len(got) != 0 {
		t.Errorf("Neighbors() = %v, want vide", IDs(got))
	}
}

func BenchmarkBacklinks(b *testing.B) {
	items := make([]KB, 200)
	for i := range items {
		target := "KB-0000"
		notes := "voir [[" + target + "]] et du texte de note."
		item, err := NewKB(
			"KB-"+string(rune('A'+i%26))+string(rune('A'+i/26)),
			"", notes, "",
		)
		if err != nil {
			b.Fatal(err)
		}
		items[i] = item
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Backlinks(items, "KB-0000")
	}
}
