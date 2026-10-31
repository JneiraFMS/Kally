package pages

import (
	"grz/html/elements"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// kbFormFields décrit le formulaire de création/modification d'une KB.
var kbFormFields = []elements.ChampFormulaire{
	{Nom: "id", Type: "text", Label: "Identifiant de la KB (ex: KB-0001) :"},
	{Nom: "erreur", Type: "textarea", Label: "Message d'erreur :"},
	{Nom: "notes", Type: "textarea", Label: "Notes :"},
	{Nom: "resolution", Type: "textarea", Label: "Message de résolution :"},
}

// PageKBCreate affiche le formulaire de création d'une KB.
func PageKBCreate(currentPath string) gomponents.Node {
	return layout("Créer une KB", currentPath, []gomponents.Node{
		Div(Class("p-8 max-w-2xl mx-auto space-y-4"),
			H1(Class("text-2xl font-bold"), gomponents.Text("Créer une KB")),

			// Le formulaire pointe vers POST /kb
			elements.FormCustomWithAction(kbFormFields, "/kb", "Enregistrer"),

			// Zone où s'affichera la réponse du serveur (htmx)
			Div(ID("reponse-serveur"), Class("text-gray-700 font-semibold")),
		),
	})
}
