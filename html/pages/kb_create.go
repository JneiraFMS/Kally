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
	{Nom: "notes", Type: "textarea", Label: "Notes (markdown) :", Rows: 6, Images: true},
	{Nom: "resolution", Type: "textarea", Label: "Résolution (markdown) :", Rows: 6, Images: true},
	{Nom: "tags", Type: "text", Label: "Tags (séparés par espace ou virgule) :"},
}

// kbQuickFormFields décrit le formulaire de création libre : un identifiant
// saisi à la main, une seule note markdown libre et des tags optionnels.
// Erreur et Résolution sont absentes (elles restent complètables via le
// formulaire d'édition complet).
var kbQuickFormFields = []elements.ChampFormulaire{
	{Nom: "id", Type: "text", Label: "Identifiant de la KB (ex: KB-0001) :"},
	{Nom: "notes", Type: "textarea", Label: "Note libre (markdown) :", Rows: 14, Images: true},
	{Nom: "tags", Type: "text", Label: "Tags (séparés par espace ou virgule) :"},
}

// FormulaireKB transporte l'état d'un formulaire de KB rendu après une
// erreur : valeurs déjà saisies (clé = name du champ) et message à afficher
// dans la zone #reponse-serveur. Les champs en lecture seule ne sont jamais
// écrasés par Valeurs.
type FormulaireKB struct {
	Valeurs map[string]string
	Message gomponents.Node
}

// PageKBCreate affiche le formulaire de création d'une KB.
func PageKBCreate(currentPath string) gomponents.Node {
	return PageKBCreateWith(currentPath, FormulaireKB{})
}

// PageKBCreateWith affiche le formulaire de création en conservant les
// valeurs saisies et le message de validation (flux sans JavaScript). La
// bascule d'enregistrement local permet de viser le dossier utilisateur
// (script grz-local) au lieu du serveur.
func PageKBCreateWith(currentPath string, form FormulaireKB) gomponents.Node {
	return kbFormPage(currentPath, "Créer une KB", "/kb",
		withValues(kbFormFields, form.Valeurs), form.Message, LocalSaveToggle())
}

// PageKBCreateQuick affiche le formulaire de création libre : identifiant
// saisi à la main, une seule note markdown libre, tags optionnels.
func PageKBCreateQuick(currentPath string) gomponents.Node {
	return PageKBCreateQuickWith(currentPath, FormulaireKB{})
}

// PageKBCreateQuickWith ré-affiche le formulaire libre en conservant les
// valeurs saisies et le message de validation (flux sans JavaScript). La
// bascule d'enregistrement local est également proposée.
func PageKBCreateQuickWith(currentPath string, form FormulaireKB) gomponents.Node {
	return kbFormPage(currentPath, "Nouvelle KB libre", "/kb/quick",
		withValues(kbQuickFormFields, form.Valeurs), form.Message, LocalSaveToggle())
}

// withValues retourne une copie des champs où les valeurs fournies
// remplacent celles d'origine (hors champs en lecture seule).
func withValues(base []elements.ChampFormulaire, valeurs map[string]string) []elements.ChampFormulaire {
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

// kbFormPage assemble une page de formulaire de KB : titre, formulaire et
// zone de réponse (message éventuel). Partagé par la création et l'édition
// pour que les deux formulaires restent identiques ; les nœuds extra
// (variadiques) portent la bascule d'enregistrement local (création
// seulement, jamais l'édition d'une KB partagée).
func kbFormPage(currentPath, titre, action string, champs []elements.ChampFormulaire, message gomponents.Node, extra ...gomponents.Node) gomponents.Node {
	return formPage(currentPath, titre, action, champs, message, extra...)
}

// formPage assemble une page de formulaire générique (KB comme journal) :
// titre, formulaire pointant vers action (POST natif + htmx) et zone de
// réponse où s'affiche le message du serveur.
func formPage(currentPath, titre, action string, champs []elements.ChampFormulaire, message gomponents.Node, extra ...gomponents.Node) gomponents.Node {
	var reponse []gomponents.Node
	if message != nil {
		reponse = append(reponse, message)
	}

	return layout(titre, currentPath, []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6"),
			Div(Class("max-w-2xl space-y-5"),
				H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text(titre)),

				elements.FormCustomWithActionExtra(champs, action, "Enregistrer", extra...),

				// Zone où s'affichera la réponse du serveur (htmx)
				Div(ID("reponse-serveur"), Class("text-g-text font-semibold"), gomponents.Group(reponse)),

				// Collage / glisser-déposer d'images sur les textareas markdown.
				imagePasteScriptNode(),
			),
		),
	})
}
