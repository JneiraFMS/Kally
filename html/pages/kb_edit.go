package pages

import (
	"grz/html/elements"
	kb "grz/internal/domain/kb"
	"strings"

	"maragu.dev/gomponents"
)

// PageKBEdit affiche le formulaire de modification d'une KB, pré-rempli.
func PageKBEdit(currentPath string, k kb.KB) gomponents.Node {
	return PageKBEditWith(currentPath, k.ID, FormulaireKB{Valeurs: kbValues(k)})
}

// PageKBEditWith affiche le formulaire de modification en conservant les
// valeurs saisies et le message de validation (flux sans JavaScript).
func PageKBEditWith(currentPath, id string, form FormulaireKB) gomponents.Node {
	return kbFormPage(currentPath, "Modifier "+id, "/kb/edit",
		kbEditFieldsFor(id, form.Valeurs), form.Message)
}

// kbEditFields construit les champs pré-remplis du formulaire d'édition,
// avec la même configuration que la création.
func kbEditFields(k kb.KB) []elements.ChampFormulaire {
	return withValues(kbEditFieldsFor(k.ID, nil), kbValues(k))
}

// kbEditFieldsFor décrit le formulaire d'édition : la liste dérive de celle
// de la création pour que toute option (rows, mono, images collées) reste
// identique entre les deux formulaires. Seule la lecture de l'identifiant
// change (champ verrouillé), les autres valeurs partent des valeurs fournies.
func kbEditFieldsFor(id string, valeurs map[string]string) []elements.ChampFormulaire {
	champs := withValues(kbFormFields, valeurs)
	for i := range champs {
		if champs[i].Nom != "id" {
			continue
		}
		champs[i].Valeur = id
		champs[i].Label = "Identifiant de la KB :"
		champs[i].LectureSeule = true
	}
	return champs
}

// kbValues regroupe les valeurs éditables d'une KB, indexées par name.
func kbValues(k kb.KB) map[string]string {
	return map[string]string{
		"erreur":     k.Erreur,
		"notes":      k.Notes,
		"resolution": k.Resolution,
		"tags":       strings.Join(k.Tags, " "),
	}
}
