package media

import (
	"errors"
	"strings"
	"testing"
)

// pngContent commence par la signature PNG : suffisant pour les tests du
// domaine, qui ne détectent pas le type eux-mêmes.
var pngContent = append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, []byte("payload")...)

func TestNameForIsContentAddressed(t *testing.T) {
	name, err := NameFor("image/png", pngContent)
	if err != nil {
		t.Fatalf("NameFor(): %v", err)
	}
	if !ValidName(name) {
		t.Errorf("NameFor() = %q, nom invalide", name)
	}
	if !strings.HasSuffix(name, ".png") {
		t.Errorf("NameFor() = %q, want suffixe .png", name)
	}

	again, err := NameFor("image/png", pngContent)
	if err != nil {
		t.Fatalf("NameFor() second appel: %v", err)
	}
	if again != name {
		t.Errorf("NameFor() non déterministe: %q != %q", again, name)
	}

	other, err := NameFor("image/png", append(append([]byte(nil), pngContent...), 'x'))
	if err != nil {
		t.Fatalf("NameFor() contenu différent: %v", err)
	}
	if other == name {
		t.Errorf("NameFor() = %q pour deux contenus différents", other)
	}
}

func TestNameForAcceptsKnownFormats(t *testing.T) {
	tests := []struct {
		mime string
		ext  string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"IMAGE/PNG", ".png"},
		{"image/jpeg; charset=binary", ".jpg"},
		{" image/png ", ".png"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			name, err := NameFor(tt.mime, pngContent)
			if err != nil {
				t.Fatalf("NameFor(%q): %v", tt.mime, err)
			}
			if !strings.HasSuffix(name, tt.ext) {
				t.Errorf("NameFor(%q) = %q, want suffixe %s", tt.mime, name, tt.ext)
			}
		})
	}
}

func TestNameForRejectsUnsupportedTypes(t *testing.T) {
	for _, mime := range []string{
		"image/svg+xml", // script embarqué : jamais accepté
		"application/pdf",
		"application/octet-stream",
		"text/plain",
		"",
	} {
		t.Run(mime, func(t *testing.T) {
			_, err := NameFor(mime, pngContent)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("NameFor(%q) error = %v, want ErrUnsupportedType", mime, err)
			}
		})
	}
}

func TestNameForRejectsEmptyAndOversizedContent(t *testing.T) {
	if _, err := NameFor("image/png", nil); !errors.Is(err, ErrEmpty) {
		t.Errorf("NameFor(vide) error = %v, want ErrEmpty", err)
	}

	if _, err := NameFor("image/png", make([]byte, MaxSize)); err != nil {
		t.Errorf("NameFor(taille max) error = %v, want nil", err)
	}

	_, err := NameFor("image/png", make([]byte, MaxSize+1))
	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("NameFor(trop gros) error = %v, want ErrTooLarge", err)
	}
}

func TestValidName(t *testing.T) {
	valid := strings.Repeat("a", 64) + ".png"
	if !ValidName(valid) {
		t.Errorf("ValidName(%q) = false, want true", valid)
	}

	tests := []string{
		"",
		"../x.png",
		strings.Repeat("A", 64) + ".png",   // hexadécimal en minuscules uniquement
		strings.Repeat("a", 63) + ".png",   // hash trop court
		strings.Repeat("a", 65) + ".png",   // hash trop long
		strings.Repeat("a", 64) + ".svg",   // extension refusée
		strings.Repeat("a", 64),            // extension absente
		strings.Repeat("a", 64) + ".png/x", // traversée de chemin
	}
	for _, name := range tests {
		if ValidName(name) {
			t.Errorf("ValidName(%q) = true, want false", name)
		}
	}
}

func TestContentType(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{"png", "image/png"},
		{"jpg", "image/jpeg"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
	}
	for _, tt := range tests {
		name := strings.Repeat("b", 64) + "." + tt.ext
		got, err := ContentType(name)
		if err != nil {
			t.Fatalf("ContentType(%q): %v", name, err)
		}
		if got != tt.want {
			t.Errorf("ContentType(%q) = %q, want %q", name, got, tt.want)
		}
	}

	if _, err := ContentType("../secret.png"); !errors.Is(err, ErrInvalidName) {
		t.Errorf("ContentType(invalide) error = %v, want ErrInvalidName", err)
	}
}
