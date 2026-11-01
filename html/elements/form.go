package elements

import (
	"strconv"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ChampFormulaire décrit un champ d'un formulaire.
type ChampFormulaire struct {
	Nom          string // Ex: "prenom" (pour le name="" du HTML)
	Type         string // Ex: "text", "email", "textarea"
	Label        string // Ex: "Votre prénom :"
	Valeur       string // Valeur pré-remplie (input) ou contenu (textarea)
	LectureSeule bool   // Champ non modifiable (ex: ID en modification)
	Mono         bool   // Police monospace (champs de traces / commandes)
	Rows         int    // Hauteur du textarea (0 → défaut 4 lignes)
}

// inputClass est la classe commune des champs de formulaire :
// sans bordure (l'élévation vient du fond bloc), coins arrondis, focus discret.
const inputClass = "mt-1 block w-full px-3.5 py-2.5 bg-g-block rounded-lg text-g-hi text-sm placeholder-g-muted focus:outline-none focus:ring-2 focus:ring-g-sep"

func InputCustom(champ ChampFormulaire) gomponents.Node {
	label := Label(
		Class("block text-g-base text-xs font-medium"),
		gomponents.Text(champ.Label),
	)
	readOnly := gomponents.If(champ.LectureSeule,
		gomponents.Group([]gomponents.Node{
			gomponents.Attr("readonly"),
			Class("bg-g-bg text-g-subdue cursor-not-allowed"),
		}),
	)

	if champ.Type == "textarea" {
		cls := inputClass
		if champ.Mono {
			cls += " font-mono"
		}
		rows := champ.Rows
		if rows <= 0 {
			rows = 4
		}
		return Div(
			Class("space-y-1"),
			label,
			gomponents.Group([]gomponents.Node{
				Textarea(
					Name(champ.Nom),
					ID("field-"+champ.Nom),
					gomponents.Attr("rows", strconv.Itoa(rows)),
					Class(cls),
					readOnly,
					gomponents.Text(champ.Valeur),
				),
			}),
		)
	}

	return Div(
		Class("space-y-1"),
		label,
		Input(
			Type(champ.Type),
			Name(champ.Nom),
			ID("field-"+champ.Nom),
			Value(champ.Valeur),
			Class(inputClass),
			readOnly,
		),
	)
}

// formNode factorise la construction du formulaire HTMX.
func formNode(fields []ChampFormulaire, action string, bouton string) gomponents.Node {
	var elements []gomponents.Node

	for _, champ := range fields {
		elements = append(elements, InputCustom(champ))
	}

	elements = append(elements, BoutonSubmit(bouton))

	return Form(
		// Fallback natif : sans JavaScript, le POST part quand même vers
		// action (htmx l'intercepte sinon).
		Action(action),
		Method("post"),
		// Attributs HTMX
		gomponents.Attr("hx-post", action),
		gomponents.Attr("hx-target", "#reponse-serveur"),
		// Anti double-submit : htmx désactive le bouton pendant l'échange.
		gomponents.Attr("hx-disabled-elt", "find button"),
		Class("space-y-4"),

		gomponents.Group(elements),
	)
}

// FormCustomWithAction construit un formulaire HTMX pointant vers action.
func FormCustomWithAction(fields []ChampFormulaire, action string, bouton string) gomponents.Node {
	return formNode(fields, action, bouton)
}
