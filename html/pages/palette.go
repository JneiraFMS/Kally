package pages

// Palette Solarized (mode dark, base03→base1), source de vérité unique des
// couleurs de l'application :
//   - tailwind.config.js génère les tokens g-* (les valeurs des classes
//     utilitaires dans le Go doivent rester synchronisées avec ces hex) ;
//   - les styles SVG du graphe (kb_graph.go) utilisent ces constantes Go ;
//   - les couleurs d'accent clair de Solarized (cyan #2aa198 / jaune
//     #b58900) complètent les teintes de base.
const (
	// Fond de page (metric g-bg) et texte inversé sur l'accent (g-ink).
	solarizedBase03 = "#002b36"
	// Fond des widgets/cards (g-widget : cards, navbar, blocs).
	solarizedBase02 = "#073642"
	// Fond des champs de saisie, entre widget et texte (g-block).
	solarizedBlock  = "#15414d"
	solarizedHover  = "#214c56"
	solarizedBorder = "#1b444f"
	// Filets de séparation discrets (arêtes du graphe, divide).
	solarizedSep = "#28515a"
	// Lueurs/accents délavés Solarized.
	solarizedYellow    = "#b58900"
	solarizedAccentDim = "#0c464f"
	solarizedCyan      = "#2aa198"
	// Ensemble de textes, du plus clair au plus délavé.
	solarizedBase2  = "#eee8d5" // titres (g-hi)
	solarizedBase1  = "#93a1a1" // corps (g-text)
	solarizedBase0  = "#839496" // secondaire (g-base)
	solarizedBase00 = "#657b83" // atténué (g-muted)
	solarizedBase01 = "#586e75" // très atténué (g-subdue)
	// Sémantique.
	solarizedRed   = "#dc322f"
	solarizedGreen = "#859900"
)
