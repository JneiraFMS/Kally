package pages

import (
	"strings"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBDetailPanel est le FRAGMENT servi par GET /kb/panel et injecté par
// le script du tiroir (grz-drawer) dans #kb-drawer-body : aucun layout,
// aucune navbar, aucune feuille — le shell du tiroir fournit la structure
// de page et les scripts (les <script> injectés via innerHTML ne sont pas
// exécutés, le comportement vit donc dans le shell).
//
// Le contenu réutilise exactement les builders de la page détail
// (noteSection, monoBody/mdBody, rétroliens, KB liées) : fiche adaptative
// — seules les sections non vides se rendent.
func PageKBDetailPanel(detail dto.KBDetail) gomponents.Node {
	k := detail.KB

	var sections []gomponents.Node
	addSection := func(titre, contenu string, body func(string) gomponents.Node) {
		if strings.TrimSpace(contenu) != "" {
			sections = append(sections, noteSection(titre, body(contenu)))
		}
	}
	addSection("Erreur", k.Erreur, monoBody)
	addSection("Notes", k.Notes, func(texte string) gomponents.Node {
		return mdBody(texte, detail.KnownIDs)
	})
	addSection("Résolution", k.Resolution, func(texte string) gomponents.Node {
		return mdBody(texte, detail.KnownIDs)
	})
	sections = append(sections,
		backlinksSection(detail.Backlinks),
		relatedSection(detail.Related, detail.Neighbors),
	)

	return Div(
		ID("kb-panel"),
		Class("px-6 py-5 space-y-5"),
		panelHeader(k),
		// Barre d'actions du tiroir : les mêmes boutons que la page détail
		// (Modifier + Supprimer avec confirmation htmx). La fiche reste
		// éditable sans quitter le tiroir ; le style secondaire est celui
		// de la page (bloc sans bordure, danger en rouge).
		Div(Class("flex justify-end gap-2"),
			secondaryAction("/kb/edit?id="+k.ID, "Modifier"),
			deleteButton(k.ID),
		),
		gomponents.Group(sections),
	)
}

// panelHeader : bandeau du fragment — identifiant mono (lien ServiceNow),
// module dérivé, date tabular-nums, bandeau T-codes (chips mono, notation
// /n). Sans bordure épaisse : le sous-titre vit en texte nu.
func panelHeader(k kb.KB) gomponents.Node {
	var meta []gomponents.Node
	meta = append(meta,
		H2(Class("font-mono text-lg font-semibold text-g-hi"),
			serviceNowLink(k.ID, k.ID, "hover:underline")),
	)
	if module := moduleSAP(k.Tags); module != "" {
		meta = append(meta,
			Span(Class("text-g-subdue"), gomponents.Text("·")),
			Span(Class("text-xs text-g-text"), gomponents.Text(module)),
		)
	}

	var dated []gomponents.Node
	if !k.Modified.IsZero() {
		dated = append(dated,
			Span(Class("tabular-nums font-mono text-[11px] text-g-text"),
				gomponents.Text(k.Modified.Format("2006-01-02"))),
		)
	}

	return Div(Class("space-y-2"),
		Div(Class("flex items-start justify-between gap-3 flex-wrap"),
			Div(Class("flex items-baseline gap-2 flex-wrap"), gomponents.Group(meta)),
			gomponents.Group(dated),
		),
		tcodeBanner(k),
	)
}

// tcodeBanner : chips mono des T-codes dérivés (notation /n, prêts pour
// la barre de commande SAP). Rien sans code dérivé.
func tcodeBanner(k kb.KB) gomponents.Node {
	codes := kBTcodes(k)
	if len(codes) == 0 {
		return nil
	}
	var chips []gomponents.Node
	for _, code := range codes {
		chips = append(chips, Span(
			Class("px-2 py-0.5 rounded bg-g-code border border-g-border font-mono text-[11px] text-g-accent-text"),
			gomponents.Text("/n"+code),
		))
	}
	return Div(Class("flex flex-wrap items-center gap-1.5"), gomponents.Group(chips))
}
