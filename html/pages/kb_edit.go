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
	return layout("Modifier "+k.ID, currentPath, []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6"),
			Div(Class("max-w-2xl space-y-5"),
				H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Modifier "+k.ID)),

				// Le formulaire pointe vers POST /kb/edit
				elements.FormCustomWithAction(kbEditFields(k), "/kb/edit", "Enregistrer"),

				// Zone où s'affichera la réponse du serveur (htmx) — même
				// position et mêmes classes que sur le formulaire de création.
				Div(ID("reponse-serveur"), Class("text-g-text font-semibold")),
			),
		),
	})
}

// kbEditFields construit les champs pré-remplis du formulaire d'édition,
// avec la même configuration que la création.
func kbEditFields(k kb.KB) []elements.ChampFormulaire {
	return []elements.ChampFormulaire{
		{Nom: "id", Type: "text", Label: "Identifiant de la KB :", Valeur: k.ID, LectureSeule: true},
		{Nom: "erreur", Type: "textarea", Label: "Message d'erreur :", Valeur: k.Erreur, Mono: true, Rows: 8},
		{Nom: "notes", Type: "textarea", Label: "Notes (markdown) :", Valeur: k.Notes, Rows: 6},
		{Nom: "resolution", Type: "textarea", Label: "Résolution (markdown) :", Valeur: k.Resolution, Rows: 6},
		{Nom: "tags", Type: "text", Label: "Tags (séparés par espace ou virgule) :", Valeur: strings.Join(k.Tags, " ")},
	}
}
