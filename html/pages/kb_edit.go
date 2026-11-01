package pages

import (
	"grz/html/elements"
	"grz/internal/kb"
	"strings"

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
		{Nom: "tags", Type: "text", Label: "Tags (séparés par espace ou virgule) :", Valeur: strings.Join(k.Tags, " ")},
	}

	return layout("Modifier "+k.ID, currentPath, []gomponents.Node{
		Div(Class("p-6 max-w-2xl mx-auto space-y-5"),
			H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Modifier "+k.ID)),

			// Le formulaire pointe vers POST /kb/edit
			elements.FormCustomWithAction(fields, "/kb/edit", "Enregistrer"),
		),

		// Zone où s'affichera la réponse du serveur (htmx)
		Div(ID("reponse-serveur"), Class("max-w-2xl mx-auto px-6 text-g-text font-semibold")),
	})
}
