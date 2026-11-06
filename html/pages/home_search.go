package pages

import (
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// fuzzyScript est le moteur de recherche fuzzy 100 % client, partagé par
// toutes les pages qui en déclarent un (accueil : cartes de KB ; journal :
// entrées listées). Le filtrage porte sur l'attribut data-search local,
// déjà rendu par le serveur : il est instantané, sans requête réseau ni
// re-rendu HTML — plus rapide que tout serveur.
//
// Défi d'éléments dupliqués : la page d'accueil contient DEUX blocs
// [data-search] par KB (une carte et une ligne de tableau). Chaque KB doit
// être comptée une seule fois : chaque élément est rattaché à son
// conteneur [data-fuzzy-view] (grille ou tableau) et seuls les éléments
// des vues visibles comptent dans « n/total ». Sans conteneur (journal),
// closest() retourne nil : élément compté, comportement historique inchangé.
//
// Markup déclaratif : le champ porte data-fuzzy et déclare ses acolytes :
//   - data-fuzzy-count      : id du compte « n/total »,
//   - data-fuzzy-no-results : id du bloc « aucun résultat » (caché/default),
//   - data-fuzzy-empty      : libellé du message vide (« Aucune KB… »),
//   - data-fuzzy-order      : (optionnel) re-trier par pertinence via
//     style.order — grille d'accueil ; la liste du journal reste
//     chronologique.
//
// Principe :
//   - index construit lazily au premier appui de touche (querySelectorAll
//     une seule fois, jamais re-scantré),
//   - scoreur sous-séquence : la requête n'a pas besoin d'être contiguë
//     ("kbqte" matche "quantite"), bonus lettres consécutives et débuts
//     de mot pour classer les meilleurs matches,
//   - combinable avec tout filtre serveur (?tag=…) : le script n'agit que
//     sur l'affichage des éléments déjà fournis par le serveur,
//   - raccourci "/" pour focuser le champ, Échap pour effacer.
const fuzzyScript = `/* grz-fuzzy */
(function () {
"use strict";

var inputs = document.querySelectorAll("input[data-fuzzy]");
if (!inputs.length) return;
var input = inputs[0];
var count = document.getElementById(input.getAttribute("data-fuzzy-count"));
var noResults = document.getElementById(input.getAttribute("data-fuzzy-no-results"));
var emptyLabel = input.getAttribute("data-fuzzy-empty") || "Aucun résultat pour";
var reorder = input.hasAttribute("data-fuzzy-order");

var cards = null;

// Index lazy : au premier appui de touche seulement, on capture chaque
// élément et sa clé data-search. Aucun coût au chargement de la page.
function collect() {
	cards = [];
	var nodes = document.querySelectorAll("[data-search]");
	for (var i = 0; i < nodes.length; i++) {
		var key = nodes[i].getAttribute("data-search");
		if (key) cards.push({ el: nodes[i], text: key });
	}
}

// fnFuzzyScore retourne un score > 0 si query est une sous-séquence de
// text (bonus : lettres consécutives, début de mot), -1 sinon.
// text est déjà en minuscules (attribut data-search du serveur).
function fnFuzzyScore(query, text) {
	var qi = 0, last = -2, streak = 0, score = 0;
	for (var i = 0; i < text.length && qi < query.length; i++) {
		if (text[i] !== query[qi]) continue;
		streak = (i === last + 1) ? streak + 1 : 1;
		var prev = i === 0 ? " " : text[i - 1];
		if (prev === " " || prev === "-" || prev === "_" || prev === ".") score += 15;
		score += 5 + streak * 4;
		last = i;
		qi++;
	}
	return qi < query.length ? -1 : score;
}

// inVisibleView : l'élément fait vivre la vue active (il doit être
// compté/trié) — rattachable à un conteneur data-fuzzy-view visible ou
// sans conteneur (journal : compté).
function inVisibleView(el) {
	var host = el.closest("[data-fuzzy-view]");
	return !host || !host.classList.contains("hidden");
}

function run() {
	var q = input.value.trim().toLowerCase();
	if (cards === null) collect();

	var shown = 0, total = 0;
	for (var i = 0; i < cards.length; i++) {
		var c = cards[i];
		if (!inVisibleView(c.el)) {
			// Vue inactive : l'affichage ne la court-circuite pas — caché
			// avec sa vue, remis au prochain passage.
			continue;
		}
		total++;
		var score = q === "" ? 1 : fnFuzzyScore(q, c.text);
		var show = score >= 0;
		c.el.classList.toggle("hidden", !show);
		// Tri par pertinence, seulement pour les pages qui le déclarent
		// (grille d'accueil). Les listes chronologiques sont préservées.
		if (reorder) {
			c.el.style.order = show && q !== "" ? String(1000000 - score) : "";
		}
		if (show) shown++;
	}

	if (count) count.textContent = q === "" ? "" : shown + "/" + total;
	if (noResults) {
		noResults.classList.toggle("hidden", q === "" || shown > 0);
		var msg = noResults.firstElementChild;
		if (msg) msg.textContent = emptyLabel + " « " + q + " ».";
	}
}

// "/" focuse le champ (sauf si l'utilisateur est déjà dans un champ),
// Échap efface et re-filtre.
function isEditable(el) {
	if (!el) return false;
	var tag = el.tagName;
	return tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || el.isContentEditable;
}

document.addEventListener("keydown", function (e) {
	if (e.key === "/" && !isEditable(document.activeElement) && !e.ctrlKey && !e.metaKey && !e.altKey) {
		e.preventDefault();
		input.focus();
		input.select();
	}
});

input.addEventListener("keydown", function (e) {
	if (e.key === "Escape") {
		input.value = "";
		run();
	}
});

input.addEventListener("input", run);
})();`

const viewToggleScript = `/* grz-view */
(function () {
"use strict";
var grid = document.querySelector('[data-fuzzy-view="grid"]');
var table = document.querySelector('[data-fuzzy-view="table"]');
if (!grid || !table) return;
var buttons = document.querySelectorAll("[data-view-target]");

function render(view) {
	grid.classList.toggle("hidden", view === "table");
	table.classList.toggle("hidden", view !== "table");
	for (var i = 0; i < buttons.length; i++) {
		var b = buttons[i];
		var pressed = b.getAttribute("data-view-target") === view;
		b.setAttribute("aria-pressed", pressed ? "true" : "false");
	}
}

for (var i = 0; i < buttons.length; i++) {
	buttons[i].addEventListener("click", function () {
		var view = this.getAttribute("data-view-target");
		render(view);
		try { localStorage.setItem("grz-view", view); } catch (e) {}
	});
}

// Préférence mémorisée ouverte, cartes par défaut
try {
	if (localStorage.getItem("grz-view") === "table") render("table");
} catch (e) {}
})();`

// searchField rend une recherche fuzzy déclarative style command palette :
// icône loupe à gauche, champ, compteur « n/total » et raccourci « / » à
// droite. Identique pour l'accueil et le journal : même moteur, même look.
// L'input reste le marqueur data-fuzzy du moteur (attributs inchangés).
func searchField(inputID, countID, noResultsID, placeholder, emptyLabel string, reorder bool) gomponents.Node {
	return Div(Class("relative max-w-2xl"),
		// Loupe décorative : aria-hidden, l'input porte son propre label
		searchIconSVG(),
		Input(
			ID(inputID),
			Type("search"),
			AutoComplete("off"),
			Placeholder(placeholder),
			Class("w-full pl-10 pr-24 py-2.5 rounded-xl bg-g-widget border border-g-border text-g-hi placeholder-g-text focus:border-g-cyan focus:outline-none transition"),
			gomponents.Attr("data-fuzzy"),
			gomponents.Attr("data-fuzzy-count", countID),
			gomponents.Attr("data-fuzzy-no-results", noResultsID),
			gomponents.Attr("data-fuzzy-empty", emptyLabel),
			gomponents.If(reorder, gomponents.Attr("data-fuzzy-order")),
		),
		// Accolades à droite : compteur live puis raccourci clavier
		Span(Class("absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1.5 pointer-events-none"),
			Span(ID(countID), Class("font-mono text-xs text-g-text"), gomponents.Text("")),
			gomponents.El("kbd",
				Class("rounded border border-g-border bg-g-block px-1.5 py-0.5 font-mono text-[10px] text-g-text"),
				gomponents.Text("/"),
			),
		),
	)
}

// searchIconSVG : la loupe de la command palette (décorative, posée à
// gauche du champ par la règle .fuzzy-loupe du style inline).
func searchIconSVG() gomponents.Node {
	return gomponents.El("svg",
		Class("fuzzy-loupe"),
		gomponents.Attr("viewBox", "0 0 24 24"),
		gomponents.Attr("width", "16"),
		gomponents.Attr("height", "16"),
		gomponents.Attr("aria-hidden", "true"),
		gomponents.Attr("fill", "none"),
		gomponents.El("circle",
			gomponents.Attr("cx", "11"), gomponents.Attr("cy", "11"), gomponents.Attr("r", "7"),
			gomponents.Attr("stroke", "currentColor"), gomponents.Attr("stroke-width", "2"),
		),
		gomponents.El("path",
			gomponents.Attr("d", "M20 20l-3.5-3.5"),
			gomponents.Attr("stroke", "currentColor"), gomponents.Attr("stroke-width", "2"),
			gomponents.Attr("stroke-linecap", "round"),
		),
	)
}

// noResultsBlock rend le bloc « aucun résultat » (caché par défaut, et
// positionnable par classe) dont le texte est posé par le script au
// premier filtrage : libellé + requête.
func noResultsBlock(blockID, containerClass string) gomponents.Node {
	cls := containerClass
	if cls != "" {
		cls += " "
	}
	cls += "hidden"
	return Div(ID(blockID), Class(cls),
		Div(Class("bg-g-widget rounded-xl p-10 text-center text-g-text"), gomponents.Text("")),
	)
}

// searchScriptNode retourne l'élément <script> du moteur fuzzy, à injecter
// après les éléments filtrables.
func searchScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(fuzzyScript))
}

// viewToggleScriptNode retourne l'élément <script> de la bascule vue,
// injecté avec le tableau (inutile quand il n'y a pas de cartes).
func viewToggleScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(viewToggleScript))
}
