package pages

import (
	"grz/internal/kb"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// firstLine retourne la première ligne non vide d'un texte (résumé du tableau).
func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}
func PageKBList(currentPath string, kbs []kb.KB) gomponents.Node {
	content := []gomponents.Node{
		Div(Class("p-8 max-w-3xl mx-auto space-y-4"),
			Div(Class("flex items-center justify-between"),
				H1(Class("text-2xl font-bold"), gomponents.Text("Bases de connaissance (KB)")),
				A(Href("/kb/new"),
					Class("px-4 py-2 rounded text-white bg-blue-600 hover:bg-blue-700"),
					gomponents.Text("+ Nouvelle KB"),
				),
			),
		),
	}

	if len(kbs) == 0 {
		content = append(content,
			Div(Class("p-8 max-w-3xl mx-auto"),
				P(Class("text-gray-500"), gomponents.Text("Aucune KB enregistrée. Créez-en une !")),
			),
		)
	} else {
		var rows []gomponents.Node
		for _, k := range kbs {
			rows = append(rows,
				Tr(Class("border-b border-gray-200"),
					Td(Class("px-4 py-2 font-mono"), gomponents.Text(k.ID)),
					Td(Class("px-4 py-2"), gomponents.Text(firstLine(k.Erreur))),
					Td(Class("px-4 py-2 text-right space-x-3"),
						A(Href("/kb/detail?id="+k.ID),
							Class("text-blue-600 hover:underline"), gomponents.Text("Consulter")),
						A(Href("/kb/edit?id="+k.ID),
							Class("text-gray-600 hover:underline"), gomponents.Text("Modifier")),
					),
				),
			)
		}

		content = append(content,
			Div(Class("max-w-3xl mx-auto px-8"),
				Table(Class("w-full text-left"),
					THead(
						Th(Class("px-4 py-2"), gomponents.Text("ID")),
						Th(Class("px-4 py-2"), gomponents.Text("Erreur")),
						Th(Class("px-4 py-2")),
					),
					TBody(gomponents.Group(rows)),
				),
			),
		)
	}

	return layout("Liste des KB", currentPath, content)
}
