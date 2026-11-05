package pages

import (
	"sort"
	"strings"

	"grz/html/elements"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var navElements = []elements.NavbarElements{
	{Label: "Accueil", Href: "/"},
	{Label: "Journal", Href: "/journal"},
	{Label: "Graphe", Href: "/kb/graph"},
}

// themeCSSRules construit les blocs `:root` (dark, par défaut) et
// `[data-theme="light"]` depuis les deux palettes : les tokens g-* sont des
// variables CSS, la bascule claire/sombre ne change QUE ces var (la feuille
// Tailwind générée les référence, aucune hex utile en dur). Ordre
// déterministe (clés triées) pour un rendu et des tests stables.
func themeCSSRules() string {
	themes := []struct {
		selector string
		vars     ThemeVars
		scheme   string
	}{
		{selector: ":root", vars: paletteDark, scheme: "dark"},
		{selector: `[data-theme="light"]`, vars: paletteLight, scheme: "light"},
	}
	var sb strings.Builder
	for _, th := range themes {
		sb.WriteString(th.selector + "{" + "color-scheme: " + th.scheme + ";")
		names := make([]string, 0, len(th.vars))
		for name := range th.vars {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			sb.WriteString("--" + name + ":" + th.vars[name] + ";")
		}
		sb.WriteString("}")
	}
	return sb.String()
}

// themeInitScript : anti-FOUC — applique data-theme sur <html> AVANT le
// chargement de la feuille CSS (préférence mémorisée, sinon le préférence
// système). Marqueur grz-theme-init.
const themeInitScript = `/* grz-theme-init */
(function () {
"use strict";
var t = null;
try { t = localStorage.getItem("grz-theme"); } catch (e) {}
if (t !== "light" && t !== "dark") {
	t = (window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches) ? "light" : "dark";
}
document.documentElement.dataset.theme = t;
})();`

// themeToggleScript : bascule clair/sombre au clic sur #theme-toggle
// (persisté en localStorage, appliqué sans rechargement grâce aux var CSS).
const themeToggleScript = `/* grz-theme-toggle */
(function () {
"use strict";
var button = document.getElementById("theme-toggle");
if (!button) return;
button.addEventListener("click", function () {
	var next = document.documentElement.dataset.theme === "light" ? "dark" : "light";
	document.documentElement.dataset.theme = next;
	try { localStorage.setItem("grz-theme", next); } catch (e) {}
});
})();`

// glanceBodyStyle : style global du corps (fond Solarized + typo + focus
// clavier visible) et styles markdown des fiches (.md) — consommés via les
// var(--g-*). La feuille des classes utilitaires, générée hors-ligne par
// tools/build-css.sh, est servie à /assets/tailwind.css.
const glanceBodyStyle = `html { scrollbar-gutter: stable; }
body {
	background-color: var(--g-bg);
	color: var(--g-text);
	font: 14px/1.6 ui-sans-serif, system-ui, sans-serif;
	-webkit-font-smoothing: antialiased;
}
::selection { background: rgba(178, 137, 0, 0.30); }
a:focus-visible, button:focus-visible, input:focus-visible, textarea:focus-visible {
	outline: 2px solid var(--g-cyan);
	outline-offset: 2px;
}
/* Champ signalé fautif par la validation (classe posée par kb_handler) :
   contour d'erreur thémé, prioritaire sur le contour de focus clavier. */
input.field-error, textarea.field-error,
input.field-error:focus-visible, textarea.field-error:focus-visible {
	outline: 2px solid var(--g-negative-text);
	outline-offset: 2px;
}
/* Bouton de bascule de thème, dans la navbar : la zone cliquable vient des
   classes utilitaires (px-2 py-3), la règle ne doit pas écraser le padding. */
#theme-toggle {
	color: var(--g-text);
	background: none; border: none; cursor: pointer;
	font: inherit;
}
#theme-toggle:hover { color: var(--g-hi); }
#theme-toggle:focus-visible { outline: 2px solid var(--g-cyan); outline-offset: 2px; }

/* Sections markdown des fiches (classe md) : épaisseur réduite, fonds
   blocs, code mono, liens accentués — thème via var(--g-*). */
/* Pas de débordement horizontal : les mots très longs (URL, jetons) se
   replient à la place. hérite aux balises enfants. */
.md { overflow-wrap: anywhere; }
.md p { margin: 0.25rem 0; }
.md ul, .md ol { margin: 0.25rem 0; padding-left: 1.25rem; }
.md ul { list-style: disc; }
.md ol { list-style: decimal; }
.md h1 { font-size: 1.15rem; font-weight: 600; color: var(--g-hi); margin: 0.5rem 0 0.25rem; }
.md h2 { font-size: 1.05rem; font-weight: 600; color: var(--g-hi); margin: 0.5rem 0 0.25rem; }
.md h3 { font-size: 0.95rem; font-weight: 600; color: var(--g-hi); margin: 0.5rem 0 0.25rem; }
.md code {
	font: 12px ui-monospace, SFMono-Regular, monospace;
	color: var(--g-hi); background: var(--g-code);
	padding: 0.1rem 0.3rem; border-radius: 0.25rem;
}
.md pre {
	background: var(--g-code); color: var(--g-hi);
	border: 1px solid var(--g-border);
	border-radius: 0.5rem; padding: 0.75rem 1rem; margin: 0.5rem 0;
	font: 12px/1.5 ui-monospace, SFMono-Regular, monospace;
	/* Page détail : les longues lignes se replient, jamais de barre
	   de défilement horizontale. */
	white-space: pre-wrap; overflow-wrap: anywhere;
}
.md pre code { background: none; padding: 0; color: inherit; }
.md-code-block { position: relative; }
.md-code-block pre { padding-top: 2.5rem; }
.md-code-copy {
	position: absolute; z-index: 1; top: 0.45rem; right: 0.5rem;
	border: 1px solid var(--g-border); border-radius: 0.375rem;
	min-width: 2rem; min-height: 2rem;
	display: flex; align-items: center; justify-content: center;
	padding: 0.3rem;
	background: var(--g-block); color: var(--g-text);
	font: 11px ui-sans-serif, system-ui, sans-serif;
	cursor: pointer;
}
.md-code-copy:hover { color: var(--g-hi); }
.md-code-copy:focus-visible { outline: 2px solid var(--g-cyan); outline-offset: 2px; }
.md-code-copy:disabled { cursor: wait; opacity: 0.7; }
.md a { color: var(--g-accent-text); text-decoration: underline; }
.md a.wikilink { text-underline-offset: 2px; }
.md a.wikilink-missing {
	color: var(--g-negative-text);
	text-decoration-style: dashed;
}
.md blockquote {
	border-left: 2px solid var(--g-sep); padding-left: 0.75rem;
	color: var(--g-text); margin: 0.25rem 0;
}

/* Coloration syntaxique (Chroma, mode classes) : familles token couvertes,
   le reste hérite de la couleur du texte. Tokens g-syn-* calibrés pour
   >= 4.5:1 sur g-code (les accents bruts g-cyan/g-positive/g-text
   échoueraient sur cette surface plus claire).
   NB: kd/kv/kt/kr = variantes de mots-clés ; fm = fonction/méthode ;
   ss/sa = chaîne spéciale ; mb/mx = nombres ; nv = variable. */
.chroma-k, .chroma-kd, .chroma-kn, .chroma-kp, .chroma-kr, .chroma-kv,
.chroma-kt, .chroma-ow, .chroma-kc { color: var(--g-syn-keyword); font-weight: 600; }
.chroma-s, .chroma-s1, .chroma-s2, .chroma-sd, .chroma-se, .chroma-si,
.chroma-sx, .chroma-ss, .chroma-sa, .chroma-m, .chroma-mi, .chroma-mf,
.chroma-mh, .chroma-mb, .chroma-mx, .chroma-mo, .chroma-il { color: var(--g-syn-string); }
.chroma-c, .chroma-c1, .chroma-ch, .chroma-cm, .chroma-cs, .chroma-cp,
.chroma-cd { color: var(--g-syn-comment); font-style: italic; }
.chroma-nf, .chroma-fm, .chroma-na { color: var(--g-hi); font-weight: 600; }
.chroma-nc, .chroma-nb, .chroma-no, .chroma-nn { color: var(--g-hi); font-weight: 600; }
.chroma-nv, .chroma-o { color: var(--g-hi); }
.chroma-err, .chroma-w { color: var(--g-syn-error); }`

// layout assemble la structure HTML commune (head, scripts, navbar)
// avec le thème Solarized commutable (dark par défaut).
func layout(title string, currentPath string, content []gomponents.Node) gomponents.Node {
	return components.HTML5(components.HTML5Props{
		Title:    title,
		Language: "fr",
		Head: []gomponents.Node{
			// Anti-FOUC : data-theme posé AVANT le chargement de la feuille
			gomponents.El("script", gomponents.Raw(themeInitScript)),
			// Feuille locale générée hors-ligne (tools/build-css.sh) :
			// classes utilitaires + preflight, tokens g-* = var(--g-*)
			Link(Rel("stylesheet"), Href("/assets/tailwind.css")),
			// Règles de thème (:root dark + [data-theme=light]) + body + .md
			gomponents.El("style", gomponents.Raw(themeCSSRules()+glanceBodyStyle)),
			Script(Src("https://unpkg.com/htmx.org@1.9.10")), // Pour soumettre sans recharger
		},
		Body: []gomponents.Node{
			elements.NavbarCustom(navElements, currentPath),
			// Bascule clair/sombre (script câblé après la navbar)
			gomponents.El("script", gomponents.Raw(themeToggleScript)),
			gomponents.Group(content),
		},
	})
}
