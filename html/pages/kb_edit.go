package pages

import (
	"grz/html/elements"
	"grz/internal/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBEdit affiche le formulaire de modification d'une KB, pré-rempli.
func PageKBEdit(currentPath string, k kb.KB) gomponents.Node {
	fields := []elements.ChampFormulaire{
		{Nom: "id", Type: "text", Label: "Identifiant de la KB :", Valeur: k.ID, LectureSeule: true},
		{Nom: "erreur", Type: "textarea", Label: "Message d'erreur :", Valeur: k.Erreur},
		{Nom: "notes", Type: "textarea", Label: "Notes :", Valeur: k.Notes},
		{Nom: "resolution", Type: "textarea", Label: "Message de résolution :", Valeur: k.Resolution},
	}

	return layout("Modifier "+k.ID, currentPath, []gomponents.Node{
		Div(Class("p-8 max-w-2xl mx-auto space-y-4"),
			H1(Class("text-2xl font-bold"), gomponents.Text("Modifier "+k.ID)),

			// Le formulaire pointe vers POST /kb/edit
			elements.FormCustomWithAction(fields, "/kb/edit", "Enregistrer"),
		),

		// Zone où s'affichera la réponse du serveur (htmx)
		Div(ID("reponse-serveur"), Class("max-w-2xl mx-auto px-8 text-gray-700 font-semibold")),
	})
}
