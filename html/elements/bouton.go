package elements

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// BoutonCustom : style minimaliste — accent jaune (primaire)
// ou bloc discret sans bordure (secondaire).
func BoutonCustom(texte string, estPrimaire bool) gomponents.Node {
	if estPrimaire {
		return Button(
			Type("button"),
			Class("px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 font-medium transition"),
			gomponents.Text(texte),
		)
	}
	return Button(
		Type("button"),
		Class("px-4 py-2 rounded-lg text-g-text bg-g-block hover:text-g-hi hover:bg-g-hover transition-colors"),
		gomponents.Text(texte),
	)
}

// BoutonSubmit : bouton d'envoi (accent jaune, seule couleur vive du formulaire).
func BoutonSubmit(texte string) gomponents.Node {
	return Button(
		Type("submit"),
		Class("px-4 py-2 rounded-lg text-g-ink bg-g-accent hover:brightness-110 font-medium transition"),
		gomponents.Text(texte),
	)
}
