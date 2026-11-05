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
	Images       bool   // Textarea acceptant le collage / dépôt d'images
}

// inputBaseClass : sans bordure (l'élévation vient du fond), coins arrondis,
// focus discret, placeholder assez contrasté (g-text, jamais g-muted).
const inputBaseClass = "mt-1 block w-full px-3.5 py-2.5 rounded-lg text-g-hi text-sm placeholder-g-text focus:outline-none focus:ring-2 focus:ring-g-sep"

// inputEditableClass : fond des champs modifiables.
const inputEditableClass = "bg-g-block"

// inputReadOnlyClass : un champ verrouillé reste reconnaissable (fond widget,
// texte courant, curseur interdit) sans désactiver la sélection du texte.
const inputReadOnlyClass = "bg-g-widget text-g-text cursor-not-allowed"

// inputClass retourne la classe du champ : elle est concaténée en UNE seule
// valeur car deux attributs class sur la même balise sont ignorés par le
// parseur HTML (le style lecture seule disparaîtrait).
func inputClass(champ ChampFormulaire) string {
	cls := inputBaseClass + " "
	if champ.LectureSeule {
		cls += inputReadOnlyClass
	} else {
		cls += inputEditableClass
	}
	if champ.Mono {
		cls += " font-mono"
	}
	return cls
}

// readOnlyAttr n'émet `readonly` que pour un champ verrouillé.
func readOnlyAttr(champ ChampFormulaire) gomponents.Node {
	return gomponents.If(champ.LectureSeule, gomponents.Attr("readonly"))
}

func InputCustom(champ ChampFormulaire) gomponents.Node {
	label := Label(
		For("field-"+champ.Nom),
		Class("block text-g-text text-xs font-medium"),
		gomponents.Text(champ.Label),
	)

	if champ.Type == "textarea" {
		rows := champ.Rows
		if rows <= 0 {
			rows = 4
		}
		return Div(
			Class("space-y-1"),
			label,
			Textarea(
				Name(champ.Nom),
				ID("field-"+champ.Nom),
				gomponents.Attr("rows", strconv.Itoa(rows)),
				Class(inputClass(champ)),
				readOnlyAttr(champ),
				// Marqueur lu par le script de collage (html/pages/image_paste.go).
				gomponents.If(champ.Images, gomponents.Attr("data-image-paste")),
				gomponents.Text(champ.Valeur),
			),
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
			Class(inputClass(champ)),
			readOnlyAttr(champ),
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
