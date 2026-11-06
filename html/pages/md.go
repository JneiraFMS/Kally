package pages

import (
	"bytes"
	"html"
	"log"
	"sync"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

// goldmarkOptions : rendu markdown de GRZ.
//   - highlighting : coloration syntaxique des blocs de code, en MODE
//     CLASSES (chromahtml.WithClasses + préfixe chroma-) : aucune couleur
//     inline émise — le thème actif (dark/light) colore via les var(--g-*)
//     définies dans layout.go (.chroma-*). Langages connus seulement,
//     sans détection auto (langage inconnu → bloc inerte, sûr).
//   - table : tableaux pipes GitHub (le style .md table vit dans layout.go ;
//     l'alignement demandé est rendu par un style inline de goldmark, le
//     seul cas où .md contient un attribut style).
//   - wikiLinkExtension : syntaxe [[ID]] rendue en lien vers la fiche (ou
//     vers la création préremplie pour une cible absente).
var goldmarkOptions = []goldmark.Option{
	goldmark.WithExtensions(
		extension.Table,
		highlighting.NewHighlighting(
			highlighting.WithWrapperRenderer(codeBlockWrapper),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.ClassPrefix("chroma-"),
			),
		),
		wikiLinkExtension{},
	),
}

var markdownPool = sync.Pool{
	New: func() any { return goldmark.New(goldmarkOptions...) },
}

// codeBlockWrapper entoure les blocs fenced générés par goldmark d'un bouton
// Copier. Le code inline n'emprunte pas ce renderer et reste inchangé.
// WrapperRenderer ne permet pas de remonter les erreurs d'écriture ; les
// écritures suivent le contrat de cette API goldmark-highlighting.
func codeBlockWrapper(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
	if entering {
		writeWrapper(w, `<div class="md-code-block"><button type="button" class="md-code-copy" aria-label="Copier le code" aria-live="polite" title="Copier le code"><svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="8" y="8" width="13" height="13" rx="2"></rect><path d="M16 8V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h3"></path></svg></button>`)
		if !context.Highlighted() {
			// Avec un langage inconnu, l'extension délègue le contenu brut au
			// writer Goldmark mais n'émet pas elle-même le wrapper <pre><code>.
			writeWrapper(w, `<pre><code>`)
		}
		return
	}
	if !context.Highlighted() {
		writeWrapper(w, `</code></pre>`)
	}
	writeWrapper(w, `</div>`)
}

func writeWrapper(w util.BufWriter, text string) {
	if _, err := w.WriteString(text); err != nil {
		log.Printf("render Markdown code block wrapper: %v", err)
	}
}

// codeCopyScript ajoute le comportement copier à chaque bloc fenced. Le texte
// est lu avec textContent (donc sans balises de coloration Chroma) ; la
// Clipboard API est préférée, avec execCommand comme fallback navigateur.
const codeCopyScript = `/* md-code-copy-script */
(function () {
"use strict";

function fallbackCopy(text) {
	return new Promise(function (resolve, reject) {
		var textarea = document.createElement("textarea");
		textarea.value = text;
		textarea.setAttribute("readonly", "");
		textarea.style.position = "fixed";
		textarea.style.opacity = "0";
		document.body.appendChild(textarea);
		textarea.select();
		var copied = false;
		try { copied = document.execCommand("copy"); } catch (e) {}
		document.body.removeChild(textarea);
		if (copied) resolve(); else reject(new Error("copy failed"));
	});
}

function copyText(text) {
	if (navigator.clipboard && navigator.clipboard.writeText) {
		return navigator.clipboard.writeText(text).catch(function () { return fallbackCopy(text); });
	}
	return fallbackCopy(text);
}

var copyIcon = '<svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="8" y="8" width="13" height="13" rx="2"></rect><path d="M16 8V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h3"></path></svg>';
var copiedIcon = '<svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m5 12 4 4L19 6"></path></svg>';
var copyErrorIcon = '<svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="m15 9-6 6m0-6 6 6"></path></svg>';

var buttons = document.querySelectorAll(".md-code-copy");
for (var i = 0; i < buttons.length; i++) {
	(function (button) {
		button.addEventListener("click", function () {
			var code = button.parentElement.querySelector("pre > code");
			if (!code) return;
			button.disabled = true;
			copyText(code.textContent || "").then(function () {
				button.innerHTML = copiedIcon;
				button.setAttribute("aria-label", "Code copié dans le presse-papiers");
			}).catch(function () {
				button.innerHTML = copyErrorIcon;
				button.setAttribute("aria-label", "Impossible de copier le code");
			}).then(function () {
				button.disabled = false;
				window.setTimeout(function () {
					button.innerHTML = copyIcon;
					button.setAttribute("aria-label", "Copier le code");
				}, 1800);
			});
		});
	})(buttons[i]);
}
})();`

// codeCopyScriptNode retourne le script à placer après le contenu markdown.
func codeCopyScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(codeCopyScript))
}

// mdToNode rend du markdown en HTML sûr via goldmark : le HTML brut du
// contenu est ignoré (option Unsafe désactivée, comportement par défaut),
// les enrichissements markdown restent fonctionnels (gras, listes, code).
// En cas d'erreur de rendu, le texte brut est affiché (échappé par
// gomponents.Text) plutôt qu'une fiche vide.
//
// Les liens wiki [[ID]] sont rendus comme des liens existants : utiliser
// mdToNodeWithKnown pour distinguer les cibles absentes.
func mdToNode(md string) gomponents.Node {
	return mdToNodeWithKnown(md, nil)
}

// mdToNodeWithKnown rend du markdown en résolvant les liens wiki contre les
// IDs connus. known == nil signifie « résolveur inconnu » (tous les liens
// sont considérés existants) ; une liste non nil, même vide, active la
// détection des liens morts.
func mdToNodeWithKnown(md string, known []string) gomponents.Node {
	markdown := markdownPool.Get().(goldmark.Markdown)
	var buf bytes.Buffer
	opts := []parser.ParseOption{}
	if known != nil {
		pc := parser.NewContext()
		pc.Set(knownIDsKey, known)
		opts = append(opts, parser.WithContext(pc))
	}
	err := markdown.Convert([]byte(md), &buf, opts...)
	markdownPool.Put(markdown)
	if err != nil {
		return Div(Class("md"), gomponents.Text(md))
	}
	return Div(Class("md"), gomponents.Raw(buf.String()))
}

// RenderMarkdownHTML rend un contenu markdown en fragment HTML texte :
// consommé par le script grz-local pour la fiche d'un fichier de la
// bibliothèque locale (File System Access), injecté par innerHTML. Même
// pipeline que les pages server-side (goldmark, wikilinks, chroma en
// classes) : known == nil signifie résolveur inconnu.
func RenderMarkdownHTML(md string, known []string) string {
	var out bytes.Buffer
	if err := mdToNodeWithKnown(md, known).Render(&out); err != nil {
		return `<div class="md">` + html.EscapeString(md) + `</div>`
	}
	return out.String()
}
