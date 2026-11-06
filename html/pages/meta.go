package pages

import (
	"regexp"
	"strings"

	kb "grz/internal/domain/kb"
)

// moduleSAP dérive le libellé du module SAP d'une KB depuis ses tags :
// un tag court purement alphabétique (2-4 lettres) est un code de module
// (mm, sd, fi…). Le tag générique « sap », les tags de projet (avec
// underscore), les tags longs et non alphabétiques sont ignorés ; sans
// module identifiable, la chaîne est vide (rien d'affiché, pas de fausse
// donnée).
func moduleSAP(tags []string) string {
	for _, tag := range tags {
		if tag == "sap" || strings.Contains(tag, "_") {
			continue
		}
		if len(tag) < 2 || len(tag) > 4 || !tCodeLetters.MatchString(tag) {
			continue
		}
		return "SAP " + strings.ToUpper(tag)
	}
	return ""
}

// tCodeLetters : tag module = uniquement des lettres (la validation
// domaine garantit déjà [a-z0-9_-]).
var tCodeLetters = regexp.MustCompile(`^[a-zA-Z]+$`)

// tCodePattern : transaction SAP — 2 à 4 lettres suivi de 2 chiffres et
// d'une lettre variante optionnelle (ME22N, MM02, FB60, VA01N), référencée
// en notation logique /nME22N ou brute. RE2 n'ayant pas de lookahead, la
// délimitation est vérifiée après coup (tCodeAtBoundary).
var tCodePattern = regexp.MustCompile(`(?:/n)?([A-Z]{2,4}\d{2}[A-Z]?)`)

// extractTCodes repère les codes de transaction SAP dans un bloc de
// markdown, dédupliqués dans l'ordre d'apparition, borné à 3 (au-delà, le
// bandeau perd sa lisibilité).
func extractTCodes(text string) []string {
	var codes []string
	seen := map[string]bool{}
	for _, loc := range tCodePattern.FindAllStringSubmatchIndex(text, -1) {
		if !tCodeAtBoundary(text, loc[0], loc[1]) {
			continue
		}
		code := text[loc[2]:loc[3]]
		if seen[code] {
			continue
		}
		seen[code] = true
		codes = append(codes, code)
		if len(codes) == 3 {
			break
		}
	}
	return codes
}

// tCodeAtBoundary vérifie que le code reconnu est délimité : début/fin de
// texte ou caractère non alnum, avec tolérance du préfixe /n.
func tCodeAtBoundary(text string, start, end int) bool {
	beforeOk := start == 0 || !isAlnum(text[start-1])
	if !beforeOk && start >= 2 && text[start-1] == 'n' && text[start-2] == '/' {
		// préfixe /n toléré quand le slash lui-même est délimité
		beforeOk = start == 2 || !isAlnum(text[start-3])
	}
	afterOk := end == len(text) || !isAlnum(text[end])
	return beforeOk && afterOk
}

// isAlnum est le test alnum ASCII borné du pattern (les contenus sont
// français : les accents ne forment pas de codes).
func isAlnum(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// kBTcodes retourne les T-codes dérivés d'une KB : les sections cousues
// Notes + Résolution (l'Erreur est une trace brute, jamais liée).
func kBTcodes(k kb.KB) []string {
	return extractTCodes(k.Notes + "\n" + k.Resolution)
}

// ModuleSAPFor expose la dérivation module aux adapters (bibliothèque
// locale) : un seul calcul, aucune règle dupliquée côté client — le
// serveur répond avec le libellé qu'il afficherait lui-même.
func ModuleSAPFor(tags []string) string {
	return moduleSAP(tags)
}

// TCodesFor expose l'extraction des T-codes (sections Notes + Résolution)
// aux adapters : même source de vérité que les cartes serveur.
func TCodesFor(notes, resolution string) []string {
	return extractTCodes(notes + "\n" + resolution)
}
