package pages

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"grz/internal/domain/wikilink"
)

// knownIDsKey transporte, pour une conversion donnée, l'ensemble des IDs de
// KB connus. Le contexte goldmark est propre à chaque appel de Convert : le
// pool d'instances Markdown reste donc partagé et sûr en concurrence.
//
// Absence de valeur (nil) = résolveur inconnu : tous les liens sont rendus
// comme des liens normaux (rendu générique, ex. journal). Valeur présente
// (même vide) = ensemble connu : une cible absente devient un lien mort.
var knownIDsKey = parser.NewContextKey()

// wikiLinkExtension ajoute la syntaxe [[ID]] au rendu Markdown. Le parseur
// inline produit directement un nœud Link : le renderer HTML standard de
// goldmark échappe le libellé et l'URL, aucun renderer custom n'est requis.
type wikiLinkExtension struct{}

// Extend implements goldmark.Extender.
func (wikiLinkExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		// Avant LinkParser (200) : [[...]] doit être reconnu comme lien wiki,
		// pas comme un lien Markdown dont le libellé serait "[...]".
		util.Prioritized(wikiLinkParser{}, 150),
	))
}

// wikiLinkParser reconnaît [[cible]] dans une ligne. Les blocs de code et le
// code inline ne passent jamais par ce parseur (goldmark les traite avant /
// différemment), et un crochet échappé ne déclenche pas le parseur.
type wikiLinkParser struct{}

// Trigger implements parser.InlineParser.
func (wikiLinkParser) Trigger() []byte {
	return []byte{'['}
}

// Parse implements parser.InlineParser.
func (wikiLinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 5 || line[0] != '[' || line[1] != '[' {
		return nil
	}
	// ![[...]] reste une syntaxe d'image : le lien wiki ne s'y invite pas.
	if block.PrecendingCharacter() == '!' {
		return nil
	}

	end := bytes.Index(line[2:], []byte("]]"))
	if end < 0 {
		return nil
	}
	target := wikilink.Target(strings.TrimSpace(string(line[2 : 2+end])))
	if !wikilink.ValidTarget(target) {
		return nil
	}

	block.Advance(2 + end + 2)
	return newWikiLinkNode(string(target), knownIDs(pc))
}

// knownIDs retourne l'ensemble des IDs connus du contexte, ou nil quand aucun
// résolveur n'est fourni (tous les liens sont alors considérés existants).
func knownIDs(pc parser.Context) map[string]struct{} {
	value := pc.Get(knownIDsKey)
	if value == nil {
		return nil
	}
	ids, ok := value.([]string)
	if !ok {
		return nil
	}
	known := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		known[id] = struct{}{}
	}
	return known
}

// newWikiLinkNode construit le lien : fiche de la KB quand la cible existe,
// formulaire de création prérempli quand elle est absente (lien mort).
func newWikiLinkNode(target string, known map[string]struct{}) ast.Node {
	link := ast.NewLink()
	link.AppendChild(link, ast.NewString([]byte(target)))

	if known == nil {
		link.Destination = []byte("/kb/detail?id=" + target)
		link.SetAttributeString("class", "wikilink")
		return link
	}
	if _, ok := known[target]; ok {
		link.Destination = []byte("/kb/detail?id=" + target)
		link.SetAttributeString("class", "wikilink")
		return link
	}

	link.Destination = []byte("/kb/new?id=" + target)
	link.Title = []byte("Créer la KB " + target)
	link.SetAttributeString("class", "wikilink wikilink-missing")
	return link
}
