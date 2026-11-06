package pages

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// drawerShell : structure fixe du tiroir latéral (slide-over), rendue
// fermée par l'accueil quand il y a des cartes. Le script grz-drawer
// fetch le fragment GET /kb/panel?id=… dans #kb-drawer-body, ouvre le
// tiroir, et fournit la fermeture (overlay, bouton, Échap) plus le
// comportement presse-papiers des blocs copier (les scripts du fragment
// ne s'exécutent pas via innerHTML, le comportement vit dans le shell).
//
// Amélioration progressive : les ancres cartes/lignes gardent leur href
// natif /kb/detail?id=… — sans JavaScript, la navigation fonctionne.
//
// État fermé par CSS pur : le tiroir est hors écran (translateX(100%)) et
// l'overlay transparent + insensible aux pointeurs ; la classe .open
// inverse les deux. Aucun JS requis pour ne rien voir au premier rendu.
func drawerShell() gomponents.Node {
	return gomponents.Group([]gomponents.Node{
		drawerStyleNode(),
		// Overlay : ferme le tiroir au clic (comportement scripté)
		Div(
			ID("kb-drawer-overlay"),
			Class("kb-drawer-overlay fixed inset-0 z-40 bg-black/50"),
		),
		Aside(
			ID("kb-drawer"),
			gomponents.Attr("role", "dialog"),
			gomponents.Attr("aria-modal", "true"),
			gomponents.Attr("aria-label", "Détail de la KB"),
			Class("fixed inset-y-0 right-0 z-50 w-full max-w-2xl flex flex-col bg-g-widget border-l border-g-border shadow-2xl"),
			// Bandeau : titre miniature + bouton fermer
			Div(Class("flex items-center justify-between px-6 py-3 border-b border-g-border shrink-0"),
				Span(Class("font-mono text-[11px] uppercase tracking-[0.2em] text-g-text"),
					gomponents.Text("KB")),
				Button(
					ID("kb-drawer-close"),
					Type("button"),
					gomponents.Attr("aria-label", "Fermer le détail"),
					Class("kb-drawer-close text-g-text hover:text-g-hi text-2xl leading-none px-2"),
					gomponents.Text("×"),
				),
			),
			// Corps : le fragment y est injecté (innerHTML), défilement local
			Div(ID("kb-drawer-body"), Class("flex-1 overflow-y-auto")),
		),
		gomponents.El("script", gomponents.Raw(drawerScript)),
		copyScriptNode(),
	})
}

// drawerCSS : transitions du tiroir — couleurs via tokens (pas de hex) ;
// la classe .open bascule position et opacité.
const drawerCSS = `
#kb-drawer { transition: transform 240ms cubic-bezier(.2,.8,.2,1); transform: translateX(100%); }
#kb-drawer.open { transform: translateX(0); }
#kb-drawer-overlay { transition: opacity 240ms ease; opacity: 0; }
#kb-drawer-overlay.open { opacity: 1; pointer-events: auto; }
#kb-drawer, #kb-drawer-overlay { pointer-events: none; }
#kb-drawer.open { pointer-events: auto; }
`

// DrawerStyleNode : le <style> du tiroir, consommant les tokens du thème.
func drawerStyleNode() gomponents.Node {
	return gomponents.El("style", gomponents.Raw(drawerCSS))
}

// drawerScript : le comportement du tiroir — ouverture via les ancres
// data-kb-drawer (fetch du fragment, repli natif sur le href si le réseau
// échoue), fermeture par overlay / bouton / Échap, restitution du focus.
const drawerScript = `/* grz-drawer */
(function () {
"use strict";
var drawer = document.getElementById("kb-drawer");
var body = document.getElementById("kb-drawer-body");
var overlay = document.getElementById("kb-drawer-overlay");
var closeBtn = document.getElementById("kb-drawer-close");
if (!drawer || !body || !overlay || !closeBtn) return;

var lastTrigger = null;

function openDrawer(trigger) {
\tlastTrigger = trigger || null;
\tbody.scrollTop = 0;
\toverlay.classList.add("open");
\tdrawer.classList.add("open");
\tcloseBtn.focus();
\t// verrouille le défilement de la page derrière le tiroir
\tdocument.body.style.overflow = "hidden";
}

function closeDrawer() {
\tif (!drawer.classList.contains("open")) return;
\toverlay.classList.remove("open");
\tdrawer.classList.remove("open");
\tdocument.body.style.overflow = "";
\tbody.innerHTML = "";
\tif (lastTrigger) lastTrigger.focus();
}

// Ouverture : chaque ancre data-kb-drawer fetch son fragment ; en cas
// d'échec (hors-ligne, serveur indisponible), la navigation native du
// href reste la porte de secours.
document.addEventListener("click", function (e) {
\tvar anchor = e.target.closest("a[data-kb-drawer]");
\tif (!anchor) return;
\te.preventDefault();
\tvar id = anchor.getAttribute("data-kb-drawer");
\tfetch("/kb/panel?id=" + encodeURIComponent(id), { headers: { "Accept": "text/html" } })
\t\t.then(function (res) {
\t\t\tif (!res.ok) throw new Error("status " + res.status);
\t\t\treturn res.text();
\t\t})
\t\t.then(function (html) {
\t\t\tbody.innerHTML = html;
\t\t\topenDrawer(anchor);
\t\t})
\t\t.catch(function () {
\t\t\tlocation.href = anchor.getAttribute("href");
\t\t});
});

overlay.addEventListener("click", closeDrawer);
closeBtn.addEventListener("click", closeDrawer);
document.addEventListener("keydown", function (e) {
\tif ("Escape" === e.key && drawer.classList.contains("open")) closeDrawer();
});
})();`
