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
var paletteDark = ThemeVars{
	"g-bg":         solarizedBase03,
	"g-widget":     solarizedBase02,
	"g-block":      "#15414d",
	"g-hover":      "#214c56",
	"g-border":     "#1b444f",
	"g-sep":        "#28515a",
	"g-accent":     solarizedYellow,
	"g-accent-dim": "#0c464f",
	"g-cyan":       solarizedCyan,
	"g-hi":         "#eee8d5",
	"g-text":       "#93a1a1",
	"g-base":       "#839496",
	"g-muted":      "#657b83",
	"g-subdue":     "#586e75",
	"g-negative":   solarizedRed,
	"g-positive":   solarizedGreen,
	"g-ink":        solarizedBase03,
}

// paletteLight : Solarized Light — échelle inversée canonique (fond base3,
// textes base01/base00, accents gérés). Les 4 tons intermédiaires
// (g-block/hover/border/sep) sont des dérivés blanchâtres (pas de
// canonique Solarized), à valider visuellement.
var paletteLight = ThemeVars{
	"g-bg":         "#fdf6e3",
	"g-widget":     "#eee8d5",
	"g-block":      "#e9e4d0",
	"g-hover":      "#ded9c2",
	"g-border":     "#d6d0b9",
	"g-sep":        "#c8c29f",
	"g-accent":     solarizedYellow,
	"g-accent-dim": "#f5ecc9",
	"g-cyan":       solarizedCyan,
	"g-hi":         "#073642",
	"g-text":       "#586e75",
	"g-base":       "#657b83",
	"g-muted":      "#839496",
	"g-subdue":     "#93a1a1",
	"g-negative":   solarizedRed,
	"g-positive":   solarizedGreen,
	"g-ink":        "#073642",
}
