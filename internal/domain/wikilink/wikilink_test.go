package wikilink

import (
	"slices"
	"strings"
	"testing"
)

func TestParseSimple(t *testing.T) {
	got := Parse("Voir [[KB-0042]] pour le détail.")
	want := []Target{"KB-0042"}
	if !slices.Equal(got, want) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestParseMultipleKeepsOrderAndDeduplicates(t *testing.T) {
	got := Parse("[[KB-0002]] puis [[KB-0001]] puis encore [[KB-0002]].")
	want := []Target{"KB-0002", "KB-0001"}
	if !slices.Equal(got, want) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestParseTrimsTarget(t *testing.T) {
	got := Parse("[[  KB-0001  ]]")
	want := []Target{"KB-0001"}
	if !slices.Equal(got, want) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestParseIgnoresUnclosedLink(t *testing.T) {
	if got := Parse("[[KB-0001 sans fermeture"); len(got) != 0 {
		t.Errorf("Parse() = %v, want vide", got)
	}
}

func TestParseRejectsInvalidTargets(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "vide", content: "[[]]"},
		{name: "espaces internes", content: "[[KB 0001]]"},
		{name: "crochet interne", content: "[[KB[0001]]"},
		{name: "crochet fermant interne", content: "[[KB]0001]]"},
		{name: "point", content: "[[KB.0001]]"},
		{name: "slash", content: "[[KB/0001]]"},
		{name: "accent", content: "[[KBé]]"},
		{name: "trop long", content: "[[" + strings.Repeat("A", MaxTargetLength+1) + "]]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.content); len(got) != 0 {
				t.Errorf("Parse(%q) = %v, want vide", tt.content, got)
			}
		})
	}
}

func TestParseIgnoresFencedCode(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "backticks",
			content: "```\n[[KB-0001]]\n```\n[[KB-0002]]",
		},
		{
			name:    "tildes",
			content: "~~~\n[[KB-0001]]\n~~~\n[[KB-0002]]",
		},
		{
			name:    "fence avec langage et indentation",
			content: "  ```go\n  [[KB-0001]]\n  ```\n[[KB-0002]]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.content)
			want := []Target{"KB-0002"}
			if !slices.Equal(got, want) {
				t.Errorf("Parse() = %v, want %v", got, want)
			}
		})
	}
}

func TestParseIgnoresInlineCode(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "backtick simple", content: "`[[KB-0001]]` puis [[KB-0002]]"},
		{name: "double backtick", content: "``[[KB-0001]]`` puis [[KB-0002]]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.content)
			if slices.Contains(got, "KB-0001") {
				t.Errorf("Parse(%q) = %v, la cible dans le code inline est retenue", tt.content, got)
			}
		})
	}
}

// TestParseUnmatchedBacktickIsLiteral : un backtick non fermé reste du texte
// (règle CommonMark), il ne protège donc pas un [[...]] qui le suit — le
// rendu goldmark se comporte de la même façon.
func TestParseUnmatchedBacktickIsLiteral(t *testing.T) {
	got := Parse("`[[KB-0001]]")
	if !slices.Equal(got, []Target{"KB-0001"}) {
		t.Errorf("Parse() = %v, want [KB-0001] (backtick non fermé = texte)", got)
	}
}

func TestParseIgnoresEscapedOpening(t *testing.T) {
	if got := Parse(`\[[KB-0001]]`); len(got) != 0 {
		t.Errorf("Parse() = %v, want vide (ouverture échappée)", got)
	}
	if got := Parse(`\\[[KB-0001]]`); !slices.Equal(got, []Target{"KB-0001"}) {
		t.Errorf("Parse() = %v, want [KB-0001] (barre échappée, lien actif)", got)
	}
}

func TestParseDoesNotSpanLines(t *testing.T) {
	if got := Parse("[[KB-\n0001]]"); len(got) != 0 {
		t.Errorf("Parse() = %v, want vide (un lien ne traverse pas une ligne)", got)
	}
}

func TestValidTarget(t *testing.T) {
	tests := []struct {
		target Target
		want   bool
	}{
		{target: "KB-0001", want: true},
		{target: "kb_0001", want: true},
		{target: "12345", want: true},
		{target: "", want: false},
		{target: "KB 0001", want: false},
		{target: "KB/0001", want: false},
		{target: "KBé", want: false},
		{target: Target(strings.Repeat("A", MaxTargetLength)), want: true},
		{target: Target(strings.Repeat("A", MaxTargetLength+1)), want: false},
	}
	for _, tt := range tests {
		if got := ValidTarget(tt.target); got != tt.want {
			t.Errorf("ValidTarget(%q) = %v, want %v", tt.target, got, tt.want)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	content := strings.Repeat("Voir [[KB-0042]] et `[[KB-0001]]`.\n\n```go\n[[KB-9999]]\n```\n", 50)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Parse(content)
	}
}
