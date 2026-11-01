/**
 * Palette Solarized dark de GRZ — source de vérité : html/pages/palette.go
 * (tests : TestTailwindConfigMatchesPalette interdisent la dérive).
 *
 * La feuille des classes utilisées est générée hors-ligne :
 *   ./tools/build-css.sh  →  web/static/tailwind.css (embarquée via go:embed)
 * Le scan couvre les fichiers Go qui construisent les classes.
 */
module.exports = {
	content: [
		'./main.go',
		'./html/**/*.go',
		'./web/**/*.go',
	],
	theme: {
		extend: {
			colors: {
				'g-bg': '#002b36',        // base03 (fond de page)
				'g-widget': '#073642',    // base02 (cards, navbar, blocs)
				'g-block': '#15414d',     // inputs, blocs intermédiaires
				'g-hover': '#214c56',
				'g-border': '#1b444f',
				'g-sep': '#28515a',       // filets, arêtes du graphe
				'g-accent': '#b58900',    // jaune Solarized délavé
				'g-accent-dim': '#0c464f',
				'g-hi': '#eee8d5',        // base2 (titres)
				'g-text': '#93a1a1',      // base1 (corps)
				'g-base': '#839496',      // base0 (secondaire)
				'g-muted': '#657b83',     // base00 (atténué)
				'g-subdue': '#586e75',    // base01 (très atténué, décoratif)
				'g-negative': '#dc322f',  // red
				'g-positive': '#859900',  // green
				'g-ink': '#002b36',       // texte inversé sur l'accent (= fond)
			},
		},
	},
	plugins: [],
}
