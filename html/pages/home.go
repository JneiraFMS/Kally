package pages

import (
	"grz/html/elements"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var formFields = []elements.ChampFormulaire{
	{Nom: "prenom", Type: "text", Label: "Votre prénom :"},
	{Nom: "nom", Type: "text", Label: "Votre nom :"},
	{Nom: "email", Type: "email", Label: "Votre adresse email :"},
}

var navElement = []elements.NavbarElements{
	{Label: "Home", Href: "/"},
	{Label: "Contact", Href: "contact"},
	{Label: "About", Href: "about"},
}

func PageAccueil() gomponents.Node {
	return components.HTML5(components.HTML5Props{
		Title:    "Formulaire avec Go",
		Language: "fr",
		Head: []gomponents.Node{
			Script(Src("https://cdn.tailwindcss.com")),
			Script(Src("https://unpkg.com/htmx.org@1.9.10")), // Pour soumettre sans recharger
		},
		Body: []gomponents.Node{
			elements.NavbarCustom(navElement),
			Div(Class("p-8 max-w-md mx-auto space-y-4"),
				H1(Class("text-2xl font-bold"), gomponents.Text("Mon Formulaire Go")),

				// Le formulaire
				elements.FormCustom(formFields),

				// Zone où s'affichera la réponse du serveur Go
				Div(ID("reponse-serveur"), Class("text-gray-700 font-semibold")),
			),
		},
	})
}
