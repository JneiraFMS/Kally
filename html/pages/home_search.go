package pages

import (
	"maragu.dev/gomponents"
)

// fuzzyScript est le moteur de recherche fuzzy 100 % client de l'accueil.
// Pourquoi client ? Les cartes sont déjà toutes dans le DOM (rendues par
// PageAccueil) : filtrer un attribut data-search local est instantané,
// sans requête réseau ni re-rendu HTML — plus rapide que tout serveur.
//
// Principe :
//   - index construit lazily au premier appui de touche (querySelectorAll
//     une seule fois, jamais re-scantré),
//   - scoreur sous-séquence : la requête n'a pas besoin d'être contiguë
//     ("kbqte" matche "quantite"), bonus lettres consécutives et débuts
//     de mot pour classer les meilleurs matches,
//   - combinable avec le filtre ?tag= : le script n'agit que sur l'affichage
//     des cartes déjà fournies par le serveur,
//   - raccourci "/" pour focuser le champ, Échap pour effacer.
const fuzzyScript = `(function () {
"use strict";

var input = document.getElementById("kb-search");
var count = document.getElementById("kb-search-count");
var noResults = document.getElementById("kb-no-results");
if (!input) return;

var cards = null;

// Index lazy : au premier appui de touche seulement, on capture chaque
// carte et sa clé data-search. Aucun coût au chargement de la page.
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
		// Tri par pertinence : CSS order (le parent est une grille).
		c.el.style.order = show && q !== "" ? String(1000000 - score) : "";
		if (show) shown++;
	}

	if (count) count.textContent = q === "" ? "" : shown + "/" + total;
	if (noResults) {
		noResults.classList.toggle("hidden", q === "" || shown > 0);
		var msg = noResults.firstElementChild;
		if (msg) msg.textContent = "Aucune KB correspondant à « " + q + " ».";
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

// searchScriptNode retourne l'élément <script> du moteur fuzzy,
// à injecter en fin de page d'accueil (après les cartes).
func searchScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(fuzzyScript))
}
