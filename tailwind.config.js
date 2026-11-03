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
				/* Thème commutable : les tokens pointent les variables CSS
				 * générées par layout.go (:root dark + [data-theme=light]).
				 * Valeurs des palettes : html/pages/palette.go (tests
				 * TestThemePalettes + TestThemePalettesExhaustives). */
				'g-bg': 'var(--g-bg)',
				'g-widget': 'var(--g-widget)',
				'g-block': 'var(--g-block)',
				'g-hover': 'var(--g-hover)',
				'g-border': 'var(--g-border)',
				'g-sep': 'var(--g-sep)',
				'g-code': 'var(--g-code)',
				'g-accent': 'var(--g-accent)',
				'g-accent-dim': 'var(--g-accent-dim)',
				'g-accent-text': 'var(--g-accent-text)',
				'g-cyan': 'var(--g-cyan)',
				'g-hi': 'var(--g-hi)',
				'g-text': 'var(--g-text)',
				'g-base': 'var(--g-base)',
				'g-muted': 'var(--g-muted)',
				'g-subdue': 'var(--g-subdue)',
				'g-negative': 'var(--g-negative)',
				'g-negative-text': 'var(--g-negative-text)',
				'g-positive': 'var(--g-positive)',
				'g-ink': 'var(--g-ink)',
				'g-syn-keyword': 'var(--g-syn-keyword)',
				'g-syn-string': 'var(--g-syn-string)',
				'g-syn-comment': 'var(--g-syn-comment)',
				'g-syn-error': 'var(--g-syn-error)',
			},
		},
	},
	plugins: [],
}
