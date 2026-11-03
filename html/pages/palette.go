package pages

// Palette Solarized — source de vérité unique des couleurs de
// l'application :
//   - tailwind.config.js référence les variables CSS (thème commutable)
//     et les tests vérifient la sync avec paletteDark/paletteLight ;
//   - les styles inline des pages (body, .md, graphe SVG) consomment les
//     var(--g-*) définies dans layout.go (thèmeClrHelpers en tête de head).

// ThemeVars est le jeu de valeurs hex d'un thème : une entrée par token
// g-* (couverture complète, contrôlée par TestThemePalettes).
type ThemeVars map[string]string

// Hex canoniques Solarized (base dé rviveus et accents).
const (
	solarizedBase03 = "#002b36"
	solarizedBase02 = "#073642"
	solarizedRed    = "#dc322f"
	solarizedGreen  = "#859900"
	solarizedYellow = "#b58900"
	solarizedCyan   = "#2aa198"
)

// paletteDark : motif actuel (Solarized dark base03→base01).
// g-accent-text / g-negative-text sont les variantes lisibles sur les fonds
// de l'interface (>= 4.5:1) : g-accent et g-negative restent les couleurs
// décoratives (SVG, filets, halos) et ne servent pas de couleur de texte.
// g-code est la surface des balises <code>/<pre> : un cran AU-DESSUS de la
// carte (g-widget), sinon le code se confond avec elle. g-syn-* sont les
// couleurs de la coloration syntaxique, calibrées sur g-code (>= 4.5:1).
var paletteDark = ThemeVars{
	"g-bg":            solarizedBase03,
	"g-widget":        solarizedBase02,
	"g-block":         "#12373f",
	"g-hover":         "#214c56",
	"g-border":        "#1b444f",
	"g-sep":           "#28515a",
	"g-code":          "#24545e",
	"g-accent":        solarizedYellow,
	"g-accent-dim":    "#0c464f",
	"g-accent-text":   "#d7ad4a",
	"g-cyan":          solarizedCyan,
	"g-hi":            "#eee8d5",
	"g-text":          "#93a1a1",
	"g-base":          "#839496",
	"g-muted":         "#657b83",
	"g-subdue":        "#586e75",
	"g-negative":      solarizedRed,
	"g-negative-text": "#f08b88",
	"g-positive":      solarizedGreen,
	"g-ink":           solarizedBase03,
	"g-syn-keyword":   "#c3d96b",
	"g-syn-string":    "#7ed3c9",
	"g-syn-comment":   "#b9c6c6",
	"g-syn-error":     "#f6b0ad",
}

// paletteLight : Solarized Light — échelle inversée canonique (fond base3,
// textes base01/base00, accents gérés). Les 4 tons intermédiaires
// (g-block/hover/border/sep) sont des dérivés blanchâtres (pas de
// canonique Solarized), à valider visuellement.
var paletteLight = ThemeVars{
	"g-bg":            "#fdf6e3",
	"g-widget":        "#eee8d5",
	"g-block":         "#e9e4d0",
	"g-hover":         "#ded9c2",
	"g-border":        "#d6d0b9",
	"g-sep":           "#c8c29f",
	"g-code":          "#fdf6e3",
	"g-accent":        solarizedYellow,
	"g-accent-dim":    "#f5ecc9",
	"g-accent-text":   "#7a5c00",
	"g-cyan":          solarizedCyan,
	"g-hi":            "#073642",
	"g-text":          "#52686f",
	"g-base":          "#657b83",
	"g-muted":         "#839496",
	"g-subdue":        "#93a1a1",
	"g-negative":      solarizedRed,
	"g-negative-text": "#a0211e",
	"g-positive":      solarizedGreen,
	"g-ink":           solarizedBase03,
	"g-syn-keyword":   "#5c6a00",
	"g-syn-string":    "#0f6f68",
	"g-syn-comment":   "#52686f",
	"g-syn-error":     "#a0211e",
}
