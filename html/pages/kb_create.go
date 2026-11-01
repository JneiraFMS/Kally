package pages

import (
	"grz/html/elements"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// kbFormFields décrit le formulaire de création d'une KB. Notes et
// Résolution acceptent le markdown ; Erreur est une trace (mono, 8 lignes).
var kbFormFields = []elements.ChampFormulaire{
	{Nom: "id", Type: "text", Label: "Identifiant de la KB (ex: KB-0001) :"},
	{Nom: "erreur", Type: "textarea", Label: "Message d'erreur :", Mono: true, Rows: 8},
	{Nom: "notes", Type: "textarea", Label: "Notes (markdown) :", Rows: 6},
	{Nom: "resolution", Type: "textarea", Label: "Résolution (markdown) :", Rows: 6},
	{Nom: "tags", Type: "text", Label: "Tags (séparés par espace ou virgule) :"},
}

// PageKBCreate affiche le formulaire de création d'une KB.
func PageKBCreate(currentPath string) gomponents.Node {
	return layout("Créer une KB", currentPath, []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6"),
			Div(Class("max-w-2xl space-y-5"),
				H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Créer une KB")),

				// Le formulaire pointe vers POST /kb
				elements.FormCustomWithAction(kbFormFields, "/kb", "Enregistrer"),

				// Zone où s'affichera la réponse du serveur (htmx)
				Div(ID("reponse-serveur"), Class("text-g-text font-semibold")),
			),
		),
	})
}
