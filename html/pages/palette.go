package pages

// Palette Kally « console » (canon Linear / Raycast / Vercel docs) —
// source de vérité unique des couleurs de l'application :
//   - tailwind.config.js référence les variables CSS (thème commutable)
//     et les tests vérifient la sync avec paletteDark/paletteLight ;
//   - les styles inline des pages (body, .md, graphe SVG) consomment les
//     var(--g-*) définies dans layout.go (thèmeClrHelpers en tête de head).
//
// Deux thèmes partagent le MÊME jeu de tokens :
//   en sombre : fond obsidienne (#0b0f17), cartes #111726, bordures
//   slate-800, accent indigo-600 (texte accent indigo-300) ;
//   en clair  : fond slate-50, cartes blanches, bordures slate-200,
//   accent indigo-600 (texte accent indigo-700).
// Décoratif vs texte : g-accent / g-negative / g-positive sont décoratifs
// (SVG, filets, halos) ; toute COULEUR DE TEXTE passe par les variantes
// calibrées >= 4.5:1 (g-accent-text, g-negative-text, g-positive-text).

// ThemeVars est le jeu de valeurs hex d'un thème : une entrée par token
// g-* (couverture complète, contrôlée par TestThemePalettes).
type ThemeVars map[string]string

// paletteDark : motif actuel, toile obsidienne console.
// g-code est la surface des balises <code>/<pre> : un cran EN DESSOUS de la
// carte (g-widget) — canon Linear — délimitée en plus par un filet
// var(--g-border) ; g-syn-* sont les couleurs de coloration syntaxique,
// calibrées sur g-code (>= 4.5:1).
var paletteDark = ThemeVars{
	"g-bg":            "#0b0f17",
	"g-widget":        "#111726",
	"g-block":         "#151d2e",
	"g-hover":         "#1d2739",
	"g-border":        "#1e293b",
	"g-sep":           "#2a3648",
	"g-code":          "#0c1120",
	"g-accent":        "#4f46e5",
	"g-accent-dim":    "#232347",
	"g-accent-text":   "#a5b4fc",
	"g-cyan":          "#818cf8",
	"g-hi":            "#f1f5f9",
	"g-text":          "#94a3b8",
	"g-base":          "#cbd5e1",
	"g-muted":         "#64748b",
	"g-subdue":        "#475569",
	"g-negative":      "#ef4444",
	"g-negative-text": "#f87171",
	"g-positive":      "#10b981",
	"g-positive-text": "#6ee7b7",
	"g-ink":           "#ffffff",
	"g-syn-keyword":   "#a5b4fc",
	"g-syn-string":    "#6ee7b7",
	"g-syn-comment":   "#a6b5ca",
	"g-syn-error":     "#fca5a5",
}

// paletteLight : canon clair « docs » — fond slate-50, cartes blanches,
// bordures slate-200. Décoratifs émeraude/rouge ils recilent aux variantes
// texte assombries (>= 4.5:1 sur les surfaces claires).
var paletteLight = ThemeVars{
	"g-bg":            "#f8fafc",
	"g-widget":        "#ffffff",
	"g-block":         "#f1f5f9",
	"g-hover":         "#e2e8f0",
	"g-border":        "#e2e8f0",
	"g-sep":           "#cbd5e1",
	"g-code":          "#f1f5f9",
	"g-accent":        "#4f46e5",
	"g-accent-dim":    "#e0e7ff",
	"g-accent-text":   "#4338ca",
	"g-cyan":          "#4f46e5",
	"g-hi":            "#0f172a",
	"g-text":          "#475569",
	"g-base":          "#334155",
	"g-muted":         "#94a3b8",
	"g-subdue":        "#cbd5e1",
	"g-negative":      "#dc2626",
	"g-negative-text": "#b91c1c",
	"g-positive":      "#059669",
	"g-positive-text": "#065f46",
	"g-ink":           "#ffffff",
	"g-syn-keyword":   "#4338ca",
	"g-syn-string":    "#047857",
	"g-syn-comment":   "#475569",
	"g-syn-error":     "#b91c1c",
}
