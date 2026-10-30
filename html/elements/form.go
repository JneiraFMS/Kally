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

	// La boucle range sur une slice conserve l'ordre exact
	for _, champ := range fields {
		elements = append(elements, InputCustom(champ))
	}

	// Bouton d'envoi (si BoutonSubmit n'existe pas, on utilise Button directement)
	elements = append(elements, Button(
		Type("submit"),
		Class("px-4 py-2 rounded text-white bg-blue-600 hover:bg-blue-700"),
		gomponents.Text("Envoyer"),
	))

	return Form(
		// Attributs HTMX
		gomponents.Attr("hx-post", "/soumettre"),
		gomponents.Attr("hx-target", "#reponse-serveur"),
		Class("space-y-3"),

		// On injecte tous les éléments générés dynamiquement
		gomponents.Group(elements),
	)
}
