package pages

import (
	"grz/html/elements"
	domainjournal "grz/internal/domain/journal"

	"maragu.dev/gomponents"
)

// journalFormFields décrit le formulaire d'une entrée de journal : une date
// et du markdown libre — pas d'identifiant, pas de tags (l'identité de la
// note est la date).
var journalFormFields = []elements.ChampFormulaire{
	{Nom: "date", Type: "date", Label: "Date du journal :"},
	{Nom: "content", Type: "textarea", Label: "Note (markdown) :", Rows: 14},
}

// FormulaireJournal transporte l'état d'un formulaire de journal rendu
// après une erreur : valeurs déjà saisies (clé = name du champ) et message
// à afficher dans la zone #reponse-serveur.
type FormulaireJournal struct {
	Valeurs map[string]string
	Message gomponents.Node
}

// PageJournalCreateWith affiche le formulaire d'entrée du journal avec les
// valeurs fournies (la date du jour est posée par l'adapter).
func PageJournalCreateWith(currentPath string, form FormulaireJournal) gomponents.Node {
	champs := withJournalValues(journalFormFields, form.Valeurs)
	return formPage(currentPath, "Nouvelle entrée du journal", "/journal", champs, form.Message)
}

// PageJournalEdit affiche le formulaire d'édition d'une entrée : la date
// est figée (lecture seule), seul le contenu est modifié.
func PageJournalEdit(currentPath string, entry domainjournal.Entry) gomponents.Node {
	return PageJournalEditWith(currentPath, entry.Date,
		FormulaireJournal{Valeurs: map[string]string{"content": entry.Content}})
}

// PageJournalEditWith ré-affiche le formulaire d'édition en conservant les
// valeurs saisies et le message de validation (flux sans JavaScript). La
// date repart telle quelle même si elle est fautive (elle est lecture
// seule dans ce formulaire).
func PageJournalEditWith(currentPath string, date string, form FormulaireJournal) gomponents.Node {
	champs := append([]elements.ChampFormulaire(nil), journalFormFields...)
	champs[0].Valeur = date
	champs[0].LectureSeule = true
	champs = withJournalValues(champs, form.Valeurs)
	return formPage(currentPath, "Journal du "+formatDateFR(date), "/journal/edit", champs, form.Message)
}

// withJournalValues retourne une copie des champs où les valeurs fournies
// remplacent celles d'origine (hors champs en lecture seule).
func withJournalValues(base []elements.ChampFormulaire, valeurs map[string]string) []elements.ChampFormulaire {
	champs := make([]elements.ChampFormulaire, len(base))
	copy(champs, base)
	for i := range champs {
		if champs[i].LectureSeule {
			continue
		}
		if valeur, ok := valeurs[champs[i].Nom]; ok {
			champs[i].Valeur = valeur
		}
	}
	return champs
}
