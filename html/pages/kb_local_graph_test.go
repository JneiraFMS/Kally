package pages

import (
	"strings"
	"testing"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"
)

// mustKB construit une KB valide pour les tests du mini-graphe
// (signature historique : id, première ligne d'erreur, tags).
func mustKB(t *testing.T, id, erreur string, tags ...string) kb.KB {
	t.Helper()
	k, err := kb.NewKB(id, erreur, "notes", "résolution", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return k
}

// TestPageKBDetailLocalGraph : la fiche d'une KB liée embarque un
// mini-graphe local (SVG, style Logseq) : la KB au centre (accent), ses
// voisins cliquables, arêtes tagées. Sans voisin, pas de SVG.
func TestPageKBDetailLocalGraph(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "trace", "n", "r", "calcul")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	related := []kb.KB{
		mustKB(t, "KB-0002", "Autre problème", "calcul", "kpi"),
		mustKB(t, "KB-0003", "Problème connexe", "kpi"),
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k, Related: related, Neighbors: related}))

	if !strings.Contains(html, `<svg`) {
		t.Fatalf("le mini-graphe local est absent:\n%s", html)
	}
	if !strings.Contains(html, `data-kg="local"`) {
		t.Errorf("le mini-graphe local n'est pas marqué data-kg=local")
	}
	// Les voisins sont cliquables vers leur fiche
	if !strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("le voisin KB-0002 n'est pas cliquable dans le mini-graphe")
	}
	// L'arête interne subsiste (principal ↔ un voisin)
	if !strings.Contains(html, "<line") {
		t.Errorf("aucune arête rendue dans le mini-graphe")
	}
	// Le tag partagé entre voisin et principal est indiqué en tooltip
	if !strings.Contains(html, "Tag partagé") {
		t.Errorf("tooltip d'arête absent du mini-graphe")
	}
	// Le nœud central porte la classe accent
	if !strings.Contains(html, "kg-dot kg-main") {
		t.Errorf("le nœud central n'utilise pas la classe accent kg-main")
	}
}

// TestPageKBDetailLocalGraphAbsentSansVoisin : une KB isolée n'a pas de
// mini-graphe (la section « KB liées » annonce le vide).
func TestPageKBDetailLocalGraphAbsentSansVoisin(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "trace", "n", "r")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if strings.Contains(html, `data-kg="local"`) {
		t.Errorf("mini-graphe rendu alors qu'il n'y a aucun voisin:\n%s", html)
	}
}

// TestPageKBDetailLocalGraphNoNestedAnchors : les liens du mini-graphe
// (SVG <a>) ne s'imbriquent jamais.
func TestPageKBDetailLocalGraphNoNestedAnchors(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "trace", "n", "r", "calcul")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	related := []kb.KB{
		mustKB(t, "KB-0002", "Autre problème", "calcul", "kpi"),
		mustKB(t, "KB-0003", "Problème connexe", "kpi"),
	}

	noNestedAnchors(t, renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k, Related: related, Neighbors: related})))
}
