package elements

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ChampFormulaire struct {
	Nom   string // Ex: "prenom" (pour le name="" du HTML)
	Type  string // Ex: "text"
	Label string // Ex: "Votre prénom :"
}

func InputCustom(champ ChampFormulaire) gomponents.Node {
	return Div(
		Class("space-y-1"),
		Label(
			Class("block text-sm font-medium text-gray-700"),
			gomponents.Text(champ.Label), // Utilise le vrai label dynamique !
		),
		Input(
			Type(champ.Type),
			Name(champ.Nom),
			Class("mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"),
		),
	)
}

func FormCustom(fields []ChampFormulaire) gomponents.Node {
	var elements []gomponents.Node

	for _, champ := range fields {
		elements = append(elements, InputCustom(champ))
	}

	elements = append(elements, BoutonSubmit("Envoyer"))

	return Form(
		// Attributs HTMX
		gomponents.Attr("hx-post", "/soumettre"),
		gomponents.Attr("hx-target", "#reponse-serveur"),
		Class("space-y-3"),

		gomponents.Group(elements),
	)
}
