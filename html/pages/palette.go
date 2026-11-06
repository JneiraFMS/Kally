package pages

// Palette Kally « console » (canon Linear / Raycast / Vercel docs) —
// source de vérité unique des couleurs de l'application :
//   - tailwind.config.js référence les variables CSS (thème commutable)
//     et les tests vérifient la sync avec paletteDark/paletteLight ;
//   - les styles inline des pages (body, .md, graphe SVG) consomment les
//     var(--g-*) définies dans layout.go (thèmeClrHelpers en tête de head).
//
// Deux thèmes partagent le MÊME jeu de tokens :
//   en sombre : fond type Dracula (#282a36), cartes #21222c, filets
//   #44475a, accent teal #087F8C (encre claire dessus) ;
//   en clair  : fond quasi-blanc (#fcfcfd), cartes blanches, accents
//   teal/rouge/vert assombris pour tenir 4.5:1.
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
	"g-bg":            "#282a36", // Background
	"g-widget":        "#21222c", // Surface principale
	"g-block":         "#343746", // Surface secondaire
	"g-hover":         "#44475a", // Current Line / Hover
	"g-border":        "#44475a", // Border
	"g-sep":           "#6272a4", // Separator
	"g-code":          "#17181f", // Code background — couche PLUS PROFONDE que la carte
	"g-accent":        "#087F8C", // Teal (accent de remplissage/bordure)
	"g-accent-dim":    "#0c3d42", // Teal muted (décoratif)
	"g-accent-text":   "#2dd6c5", // Teal éclairci (>= 4.5:1 sur les surfaces sombres)
	"g-cyan":          "#8be9fd", // Cyan
	"g-hi":            "#f8f8f2", // Foreground
	"g-text":          "#e0e0e0", // Main text
	"g-base":          "#f8f8f2", // Base text
	"g-muted":         "#a6a6b3", // Muted text
	"g-subdue":        "#6272a4", // Subdued text
	"g-negative":      "#ff5555", // Red
	"g-negative-text": "#ff7d7d", // Light red (>= 4.5:1 sur g-block)
	"g-positive":      "#50fa7b", // Green
	"g-positive-text": "#69ff94", // Light green
	"g-ink":           "#ffffff", // Encre claire sur l'accent teal (>= 4.5:1)
	"g-syn-keyword":   "#ff79c6", // Pink
	"g-syn-string":    "#f1fa8c", // Yellow
	"g-syn-comment":   "#7c87b8", // Comment éclairci (>= 4.5:1 sur g-code)
	"g-syn-error":     "#ff5555", // Red
}

// paletteLight : canon clair « docs » — fond slate-50, cartes blanches,
// bordures slate-200. Décoratifs émeraude/rouge ils recilent aux variantes
// texte assombries (>= 4.5:1 sur les surfaces claires).
var paletteLight = ThemeVars{
	"g-bg":            "#fcfcfd",
	"g-widget":        "#ffffff",
	"g-block":         "#F0F7F7",
	"g-hover":         "#e6eff0",
	"g-border":        "#e1ebde",
	"g-sep":           "#c8d8c5",
	"g-code":          "#f0f7f7",
	"g-accent":        "#087F8C", // Teal : encre blanche >= 4.5:1
	"g-accent-dim":    "#d5ecea",
	"g-accent-text":   "#056a5f", // Teal assombri (>= 4.5:1 sur les surfaces claires)
	"g-cyan":          "#008fb2",
	"g-hi":            "#282a36",
	"g-text":          "#44475a",
	"g-base":          "#343746",
	"g-muted":         "#77778b",
	"g-subdue":        "#a6a6b3",
	"g-negative":      "#d83b52",
	"g-negative-text": "#b52b43",
	"g-positive":      "#168b55",
	"g-positive-text": "#087443",
	"g-ink":           "#ffffff",
	"g-syn-keyword":   "#a626a4",
	"g-syn-string":    "#136f34", // >= 4.5:1 sur g-code
	"g-syn-comment":   "#5f5f74", // >= 4.5:1 sur g-code
	"g-syn-error":     "#b02b41", // >= 4.5:1 sur g-code
}
