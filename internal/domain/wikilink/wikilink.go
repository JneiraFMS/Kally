// Package wikilink définit la syntaxe des liens wiki [[cible]] utilisée
// dans les notes Markdown (inspirée de Logseq) et son extraction.
//
// Le package est pur : il ne dépend ni du rendu HTML ni du stockage. La
// grammaire d'une cible est volontairement la même que celle des identifiants
// de KB (lettres, chiffres, tirets, underscores) : un lien pointe donc
// toujours vers une note identifiable.
package wikilink

import (
	"regexp"
	"strings"
)

// MaxTargetLength borne la longueur d'une cible pour éviter les liens
// accidentellement démesurés (une cible plus longue n'est pas un lien).
const MaxTargetLength = 128

// Target est la cible brute d'un lien wiki, sans les doubles crochets.
type Target string

// targetPattern est la grammaire d'une cible valide : identifiants de note
// (mêmes caractères que les IDs de KB).
var targetPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// ValidTarget reports whether t is a syntactically valid link target.
func ValidTarget(t Target) bool {
	s := string(t)
	return s != "" && len(s) <= MaxTargetLength && targetPattern.MatchString(s)
}

// Parse extrait les cibles [[...]] d'un contenu Markdown, dans l'ordre
// d'apparition et sans doublon.
//
// Les blocs de code fencés (``` ou ~~~) et le code inline (backticks) sont
// ignorés : un [[...]] qui y figure reste du texte. Une ouverture échappée
// (\[[) n'ouvre pas de lien, et un lien ne peut pas traverser une ligne.
// Les cibles invalides (vides, espaces, crochets, caractères hors grammaire)
// sont ignorées : elles restent du texte littéral.
func Parse(content string) []Target {
	targets := []Target{}
	seen := make(map[Target]struct{})

	var fenceChar byte
	var fenceLen int

	for _, line := range strings.Split(content, "\n") {
		if fenceChar != 0 {
			if closesFence(line, fenceChar, fenceLen) {
				fenceChar = 0
			}
			continue
		}
		if ch, n, ok := opensFence(line); ok {
			fenceChar, fenceLen = ch, n
			continue
		}
		scanLine(line, func(target Target) {
			if !ValidTarget(target) {
				return
			}
			if _, ok := seen[target]; ok {
				return
			}
			seen[target] = struct{}{}
			targets = append(targets, target)
		})
	}

	return targets
}

// opensFence reconnaît une ligne d'ouverture de bloc de code fencé
// (CommonMark : jusqu'à 3 espaces d'indentation, au moins 3 backticks ou
// tildes). Le reste de la ligne est l'info string éventuelle.
func opensFence(line string) (byte, int, bool) {
	i := 0
	for i < len(line) && i < 4 && line[i] == ' ' {
		i++
	}
	if i > 3 || i >= len(line) {
		return 0, 0, false
	}
	ch := line[i]
	if ch != '`' && ch != '~' {
		return 0, 0, false
	}
	n := 0
	for i < len(line) && line[i] == ch {
		n++
		i++
	}
	if n < 3 {
		return 0, 0, false
	}
	return ch, n, true
}

// closesFence reconnaît la ligne de fermeture d'un bloc fencé : même
// caractère, longueur au moins égale à l'ouverture, rien d'autre que des
// espaces ensuite.
func closesFence(line string, ch byte, minLen int) bool {
	i := 0
	for i < len(line) && i < 4 && line[i] == ' ' {
		i++
	}
	if i > 3 {
		return false
	}
	n := 0
	for i < len(line) && line[i] == ch {
		n++
		i++
	}
	if n < minLen {
		return false
	}
	for ; i < len(line); i++ {
		if line[i] != ' ' && line[i] != '\t' {
			return false
		}
	}
	return true
}

// scanLine parcourt une ligne hors bloc de code et émet chaque [[cible]]
// rencontrée hors code inline et hors échappement.
func scanLine(line string, emit func(Target)) {
	for i := 0; i < len(line); {
		switch line[i] {
		case '\\':
			// Un caractère échappé ne peut pas ouvrir un lien : on saute
			// la barre et le caractère suivant.
			if i+1 < len(line) {
				i += 2
			} else {
				i++
			}
		case '`':
			run := backtickRun(line[i:])
			if closeAt := matchingBackticks(line, i+run, run); closeAt >= 0 {
				i = closeAt
				continue
			}
			// Backtick non fermé : texte littéral, on poursuit après la
			// série de backticks.
			i += run
		case '[':
			if i+1 >= len(line) || line[i+1] != '[' {
				i++
				continue
			}
			end := strings.Index(line[i+2:], "]]")
			if end < 0 {
				i += 2
				continue
			}
			target := Target(strings.TrimSpace(line[i+2 : i+2+end]))
			emit(target)
			i += 2 + end + 2
		default:
			i++
		}
	}
}

// backtickRun retourne la longueur de la série de backticks en tête de s.
func backtickRun(s string) int {
	n := 0
	for n < len(s) && s[n] == '`' {
		n++
	}
	return n
}

// matchingBackticks cherche, à partir de start, une série de backticks de
// même longueur que run (règle CommonMark du code inline) et retourne la
// position juste après. -1 si la série n'est pas fermée.
func matchingBackticks(line string, start, run int) int {
	for i := start; i < len(line); {
		if line[i] != '`' {
			i++
			continue
		}
		n := backtickRun(line[i:])
		if n == run {
			return i + n
		}
		i += n
	}
	return -1
}
