package elements

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func BoutonCustom(texte string, estPrimaire bool) gomponents.Node {
	couleurFond := "bg-blue-600 hover:bg-blue-700"
	if !estPrimaire {
		couleurFond = "bg-gray-600 hover:bg-gray-700"
	}

	return Button(
		Type("button"),
		Class("px-4 py-2 rounded text-white "+couleurFond),
		gomponents.Text(texte),
	)
}

func BoutonSubmit(texte string) gomponents.Node {
	return Button(
		Type("submit"),
		Class("px-4 py-2 rounded text-white bg-blue-600 hover:bg-blue-700"),
		gomponents.Text(texte),
	)
}
