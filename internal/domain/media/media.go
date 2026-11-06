// Package media porte les règles des images collées dans les notes : formats
// acceptés, taille maximale et nommage adressé par le contenu. Aucune
// dépendance HTTP, disque ou framework : la détection binaire et le stockage
// appartiennent aux adapters.
package media

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// MaxSize est la taille maximale acceptée pour une image (5 Mio).
	MaxSize = 5 << 20
)

var (
	// ErrUnsupportedType est retourné quand le type MIME n'est pas une image
	// acceptée (SVG volontairement exclu : il peut embarquer du script).
	ErrUnsupportedType = errors.New("media: format non supporté")

	// ErrEmpty est retourné quand le contenu de l'image est vide.
	ErrEmpty = errors.New("media: contenu vide")

	// ErrTooLarge est retourné quand l'image dépasse MaxSize.
	ErrTooLarge = errors.New("media: image trop volumineuse")

	// ErrInvalidName est retourné quand un nom de média est mal formé.
	ErrInvalidName = errors.New("media: nom invalide")

	// ErrNotFound est retourné quand un média n'existe pas dans le stockage.
	ErrNotFound = errors.New("media: introuvable")
)

// formats associe chaque type MIME accepté à son extension de fichier.
var formats = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// mimeByExtension est l'inverse de formats (extensions uniques).
var mimeByExtension = map[string]string{
	"png":  "image/png",
	"jpg":  "image/jpeg",
	"gif":  "image/gif",
	"webp": "image/webp",
}

// namePattern : nom = sha256 hexadécimal (64 caractères) + extension connue.
var namePattern = regexp.MustCompile(`^[a-f0-9]{64}\.(png|jpg|gif|webp)$`)

// NameFor valide le type MIME et la taille, puis retourne le nom de fichier
// adressé par le contenu (<sha256>.<ext>). Deux images identiques produisent
// le même nom, ce qui rend le stockage idempotent.
func NameFor(mime string, content []byte) (string, error) {
	ext, ok := formats[normalizeMIME(mime)]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnsupportedType, mime)
	}
	if len(content) == 0 {
		return "", ErrEmpty
	}
	if len(content) > MaxSize {
		return "", fmt.Errorf("%w: %d octets", ErrTooLarge, len(content))
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]) + "." + ext, nil
}

// ValidName reports whether name is a content-addressed media name produced
// by NameFor. Anything else (path traversal, uppercase, unknown extension)
// is rejected before touching the filesystem.
func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

// ContentType retourne le type MIME associé au nom d'un média validé.
func ContentType(name string) (string, error) {
	if !ValidName(name) {
		return "", fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	mime, ok := mimeByExtension[name[strings.LastIndexByte(name, '.')+1:]]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return mime, nil
}

// normalizeMIME met le type MIME sous forme canonique : minuscules, sans
// paramètres (`image/jpeg; charset=binary` → `image/jpeg`).
func normalizeMIME(mime string) string {
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = mime[:i]
	}
	return strings.ToLower(strings.TrimSpace(mime))
}
