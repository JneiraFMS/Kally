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

function run() {
	var q = input.value.trim().toLowerCase();
	if (cards === null) collect();

	var shown = 0, total = cards.length;
	for (var i = 0; i < cards.length; i++) {
		var c = cards[i];
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
})();
`

// searchField rend une recherche fuzzy déclarative : l'input (marqueur
// data-fuzzy) et son compteur « n/total » superposé au bord droit. Identique
// pour l'accueil et le journal : même moteur, même look.
func searchField(inputID, countID, noResultsID, placeholder, emptyLabel string, reorder bool) gomponents.Node {
	return Div(Class("relative"),
		Input(
			ID(inputID),
			Type("search"),
			AutoComplete("off"),
			Placeholder(placeholder),
			Class("w-full px-4 py-2.5 rounded-lg bg-g-block border border-g-border text-g-hi placeholder-g-text focus:border-g-accent outline-none transition"),
			gomponents.Attr("data-fuzzy"),
			gomponents.Attr("data-fuzzy-count", countID),
			gomponents.Attr("data-fuzzy-no-results", noResultsID),
			gomponents.Attr("data-fuzzy-empty", emptyLabel),
			gomponents.If(reorder, gomponents.Attr("data-fuzzy-order")),
		),
		Span(ID(countID),
			Class("absolute right-3 top-1/2 -translate-y-1/2 font-mono text-xs text-g-text pointer-events-none"),
			gomponents.Text(""),
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
