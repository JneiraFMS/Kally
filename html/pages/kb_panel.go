package pages

import (
	"fmt"
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
		gomponents.Group(sections),
		// Autosuffisant : le fragment ouvert seul (navigation directe
		// /kb/panel) copie aussi ; dans le tiroir, la copie est déjà
		// fournie par le shell (guard idempotent du script)
		copyScriptNode(),
	)
}

// panelHeader : bandeau du fragment — identifiant mono (lien ServiceNow),
// module dérivé, date tabular-nums, bandeau T-codes (chips mono, notation
// /n), bloc « Copier la solution ». Sans bordure épaisse : le sous-titre
// vit en texte nu.
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
		copySolutionBlock(k),
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

// copyTargetID : id DOM de la source presse-papiers d'une fiche.
func copyTargetID(id string) string {
	return "kb-copy-" + id
}

// copySolutionBlock : le bloc « Copier la solution » — bouton +
// textarea cachée porteuse du texte prêt-à-coller (solutionText). Le
// comportement presse-papiers vit dans le script copier (inclus par le
// shell du tiroir et par la page détail : les scripts injectés via
// innerHTML ne s'exécutent pas, le markup seul se déplace).
func copySolutionBlock(k kb.KB) gomponents.Node {
	target := copyTargetID(k.ID)
	return Div(Class("flex items-center gap-2"),
		Button(
			Type("button"),
			Class("px-3 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover border border-g-border text-xs font-medium transition-colors"),
			gomponents.Attr("data-copy-target", target),
			gomponents.Attr("data-copy-done", "Copié ✓"),
			gomponents.Text("Copier la solution"),
		),
		Textarea(
			ID(target),
			Class("hidden"),
			gomponents.Attr("readonly", ""),
			gomponents.Attr("aria-hidden", "true"),
			TabIndex("-1"),
			gomponents.Text(solutionText(k)),
		),
	)
}

// solutionText compose le texte prêt-à-coller pour un ticket ServiceNow :
// en-tête fiche (ID + module), symptôme (1re ligne de l'erreur), contexte
// (notes), étapes (résolution), lien ServiceNow. Chaque bloc vide est
// omis : le texte copié reste dense et honnête.
func solutionText(k kb.KB) string {
	var b strings.Builder
	fmt.Fprintf(&b, "KB %s", k.ID)
	if module := moduleSAP(k.Tags); module != "" {
		fmt.Fprintf(&b, " — %s", module)
	}
	b.WriteString("\n")

	if titre := firstLine(k.Erreur); titre != "" {
		fmt.Fprintf(&b, "Symptôme : %s\n", titre)
	}
	if notes := strings.TrimSpace(k.Notes); notes != "" {
		fmt.Fprintf(&b, "Contexte :\n%s\n", notes)
	}
	if reso := strings.TrimSpace(k.Resolution); reso != "" {
		fmt.Fprintf(&b, "Étapes :\n%s\n", reso)
	}
	fmt.Fprintf(&b, "Référence : %s", serviceNowURL(k.ID))
	return b.String()
}

// copyScript est le comportement presse-papiers des blocs copier (tiroir
// ET page détail) : textarea cachée comme source unique de vérité,
// Clipboard API avec repli execCommand, libellé de succès restauré après
// 1500 ms (data-copy-done). Le garde-fou au niveau window rend l'écoute
// idempotente : le fragment en embarque une copie (lisible hors tiroir)
// et le shell du tiroir en embarque une (les scripts injectés via
// innerHTML ne s'exécutent pas), sans jamais doubler le handler.
const copyScript = `/* grz-copy */
(function () {
"use strict";
if (window.grzCopyBound) return;
window.grzCopyBound = true;
document.addEventListener("click", function (e) {
\tvar btn = e.target.closest("[data-copy-target]");
\tif (!btn) return;
\tvar ta = document.getElementById(btn.getAttribute("data-copy-target"));
\tif (!ta) return;

\tvar initial = btn.textContent;
\tvar done = function () {
\t\tbtn.textContent = btn.getAttribute("data-copy-done") || "Copié ✓";
\t\tsetTimeout(function () { btn.textContent = initial; }, 1500);
\t};
\tvar fallback = function () {
\t\tta.focus();
\t\tta.select();
\t\ttry { document.execCommand("copy"); } catch (err) {}
\t\tdone();
\t};
\tif (navigator.clipboard && navigator.clipboard.writeText) {
\t\tnavigator.clipboard.writeText(ta.value).then(done, fallback);
\t} else {
\t\tfallback();
\t}
});
})();`

// copyScriptNode retourne l'élément <script> du comportement copier.
func copyScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(copyScript))
}
