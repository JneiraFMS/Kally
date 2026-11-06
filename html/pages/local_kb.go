package pages

import (
	. "maragu.dev/gomponents/html"

	"maragu.dev/gomponents"
)

// Bibliothèque locale (File System Access) : le script grz-local permet à
// un utilisateur de l'app hébergée de stocker des articles en VRAIS
// fichiers .md dans un dossier de sa propre machine, consultables
// uniquement par lui. Le serveur ne persiste rien sur ce périmètre — il ne
// fournit que des transformations sans état
// (POST /kb/local/serialize | parse | render) autour du codec unique
// internal/kbfile : le fichier local suit exactement le format serveur.
//
// Monde étanche : une fiche locale n'apparaît jamais dans les listes, le
// graphe ou les rétroliens des KB partagées ; les wikilinks d'un fichier
// local peuvent viser une ID partagée (rendu avec les IDs connus du
// serveur), jamais l'inverse.
//
// Amélioration progressive :
//   - tout ce qui concerne le local porte la classe « hidden » côté
//     serveur, révélée seulement quand window.showDirectoryPicker existe
//     (Chrome/Edge) — sans JavaScript (ou Firefox/Safari) tout fonctionne
//     comme avant, simplement sans périmètre local ;
//   - le FileSystemDirectoryHandle est persisté dans IndexedDB
//     (grz-local-kb) et revalidé par queryPermission à chaque session :
//     la reconnexion explicite (clic) remplace la permission silencieuse,
//     conformément à l'API.
const localKBScript = `/* grz-local */
(function () {
"use strict";

var toggle = document.getElementById("kb-local-save");
var statusEl = document.getElementById("local-kb-status");
var connectButtons = document.querySelectorAll("[data-local-connect]");
var gridEl = document.getElementById("kb-cards");          // grille partagée, sortie des articles locaux
var tableRowsEl = document.getElementById("kb-table-rows"); // vue tableau, lignes locales
var emptyEl = document.getElementById("kb-empty");          // état vide partagé
var hasUI = !!(toggle || statusEl || connectButtons.length || gridEl);
if (!hasUI) return;
if (typeof window.showDirectoryPicker !== "function" || !window.isSecureContext) {
	// Pas de silence : la barre est révélée avec l'explication. Deux cas
	// existent : navigateur sans File System Access (Firefox/Safari/IE)
	// ou origine non sécurisée (File System Access exige HTTPS ou
	// http://localhost — un accès via http://<ip-du-serveur> est refusé
	// par les navigateurs). Le bouton de connexion disparaît, la fiche
	// explique pourquoi rien ne peut s'afficher.
	var fallbackHost = toggle ? toggle.closest("[data-local-block]") : document.getElementById("local-kb-bar");
	if (fallbackHost) fallbackHost.classList.remove("hidden");
	// Sur les formulaires, la bascule locale n'a aucun sens sans FSA : on
	// masque aussi le libellé pour ne pas laisser une case morte.
	if (toggle) {
		var toggleLabel = toggle.closest("label");
		if (toggleLabel) toggleLabel.classList.add("hidden");
	}
	var unsupportedButtons = document.querySelectorAll("[data-local-connect]");
	for (var ui = 0; ui < unsupportedButtons.length; ui++) unsupportedButtons[ui].classList.add("hidden");
	if (!window.isSecureContext) {
		setMessage("Articles locaux indisponibles : connexion sécurisée requise (accédez à l'app via https://… ou http://localhost).");
	} else {
		setMessage("Articles locaux indisponibles : Chrome ou Edge requis (File System Access).");
	}
	return;
}

var dir = null;    // FileSystemDirectoryHandle du dossier choisi
var entries = [];  // { id, parsed, modified }

// ---- handle persisté (IndexedDB) --------------------------------------
var DB_NAME = "grz-local-kb", DB_STORE = "handles", DB_KEY = "library";

function openDB() {
	return new Promise(function (resolve, reject) {
		var req = indexedDB.open(DB_NAME, 1);
		req.onupgradeneeded = function () {
			req.result.createObjectStore(DB_STORE);
		};
		req.onsuccess = function () { resolve(req.result); };
		req.onerror = function () { reject(req.error || new Error("indexedDB")); };
	});
}

function readStoredHandle() {
	return openDB().then(function (db) {
		return new Promise(function (resolve, reject) {
			var req = db.transaction(DB_STORE).objectStore(DB_STORE).get(DB_KEY);
			req.onsuccess = function () { resolve(req.result || null); };
			req.onerror = function () { reject(req.error); };
		});
	}).catch(function () { return null; });
}

function storeHandle(handle) {
	return openDB().then(function (db) {
		return new Promise(function (resolve, reject) {
			var tx = db.transaction(DB_STORE, "readwrite");
			tx.objectStore(DB_STORE).put(handle, DB_KEY);
			tx.oncomplete = function () { resolve(); };
			tx.onerror = function () { reject(tx.error); };
		});
	}).catch(function () {});
}

// queryPermission ne demande rien ; requestPermission exige un geste
// utilisateur (clic sur « Connecter »).
function granted(handle) {
	return handle.queryPermission({ mode: "readwrite" }).then(function (state) {
		return state === "granted";
	});
}

function setMessage(text) {
	if (!statusEl) return;
	statusEl.textContent = text || "";
}

// ---- connexion du dossier ----------------------------------------------
function connect() {
	return window.showDirectoryPicker({ id: "grz-local-kb", mode: "readwrite", startIn: "documents" })
		.then(function (handle) {
			return granted(handle).then(function (ok) {
				if (!ok) { setMessage("Autorisation refusée pour ce dossier."); return; }
				dir = handle;
				return storeHandle(dir).then(function () {
					revealBlock();
					if (gridEl) refreshIndex();
					else setMessage("Dossier connecté : " + dir.name);
				});
			});
		})
		.catch(function (error) {
			if (error && error.name === "AbortError") return; // annulation utilisateur
			setMessage("Connexion impossible : " + ((error && error.message) || error));
		});
}

for (var bi = 0; bi < connectButtons.length; bi++) {
	(function (button) {
		button.addEventListener("click", function () { connect(); });
	})(connectButtons[bi]);
}

// Découverte au chargement : un handle existant sans permission fraîche
// attend son clic — jamais de re-demande silencieuse.
var discovery = readStoredHandle().then(function (handle) {
	if (!handle) {
		// Première utilisation : on explique ce qui donne les articles.
		if (!toggle) setMessage("Aucun dossier connecté : cliquez sur « Connecter mon dossier » pour afficher vos articles.");
		return;
	}
	return granted(handle).then(function (ok) {
		if (ok) {
			dir = handle;
			if (gridEl) refreshIndex();
			else setMessage("Dossier connecté : " + dir.name);
		} else if (!toggle) {
			setMessage("Dossier connu mais autorisation expirée : reconnectez-le.");
		}
	});
}).catch(function () {});

function revealBlock() {
	var host = toggle ? toggle.closest("[data-local-block]") : document.getElementById("local-kb-bar");
	if (host) host.classList.remove("hidden");
}

discovery.then(function () { revealBlock(); });

// ---- index des fichiers locaux -----------------------------------------
function firstLine(text) {
	if (!text) return "";
	var lines = text.split("\n");
	for (var i = 0; i < lines.length; i++) {
		var line = lines[i].trim();
		if (line) return line;
	}
	return "";
}

// afterFirstLine / truncateTo : clones des dérivations serveur
// (afterFirstLine, excerpt) pour l'extrait de carte.
function afterFirstLine(text) {
	if (!text) return "";
	var lines = text.split("\n");
	var seen = false;
	for (var i = 0; i < lines.length; i++) {
		if (lines[i].trim()) {
			if (seen) return lines.slice(i).join("\n");
			seen = true;
		}
	}
	return "";
}

function truncateTo(text, max) {
	if (!text || text.length <= max) return text || "";
	return text.slice(0, max) + "...";
}

function headline(e) {
	return firstLine(e.parsed.erreur) || firstLine(e.parsed.notes) || firstLine(e.parsed.resolution);
}

function searchKey(e) {
	return (e.id + " " + headline(e) + " " + (e.parsed.tags || []).join(" ")).toLowerCase();
}

function dateOf(e) {
	if (!e.modified) return "";
	return new Date(e.modified).toISOString().slice(0, 10);
}

function el(tag, className, text) {
	var node = document.createElement(tag);
	if (className) node.className = className;
	if (text) node.textContent = text;
	return node;
}

function refreshIndex() {
	if (!dir) return Promise.resolve();
	entries = [];
	var jobs = [];
	var iterator = dir.values();
	function walk() {
		return iterator.next().then(function (result) {
			if (result.done) return null;
			var handle = result.value;
			var name = handle.name;
			if (handle.kind !== "file" || !/\.md$/.test(name) || name.indexOf("journal-") === 0) {
				return walk();
			}
			jobs.push(loadEntry(name.slice(0, -3), handle));
			return walk();
		});
	}
	return walk().then(function () {
		return Promise.all(jobs);
	}).then(function () {
		entries.sort(function (a, b) { return b.modified - a.modified; });
		renderLocal();
		if (gridEl) {
			var count = entries.length ? " — " + entries.length + " article(s)" : " — aucun article";
			setMessage("Dossier connecté : " + dir.name + count);
		}
	}).catch(function (error) {
		setMessage("Lecture du dossier impossible : " + ((error && error.message) || error));
	});
}

async function loadEntry(id, handle) {
	var file = await handle.getFile();
	var raw = await file.text();
	var response = await fetch("/kb/local/parse", {
		method: "POST",
		headers: { "Content-Type": "text/markdown; charset=utf-8" },
		body: raw
	});
	if (!response.ok) return;
	var parsed = await response.json();
	entries.push({ id: id, parsed: parsed, modified: file.lastModified });
}

function tagChip(tag) {
	return el("span", "px-2 py-0.5 rounded bg-g-block font-mono text-[11px] text-g-text", tag);
}

// renderLocal injecte les articles locaux DANS la grille partagée
// (#kb-cards) et la vue tableau (#kb-table-rows), à côté des KB serveur :
// même canevas que kbCard, comptés dans « n/total » du moteur fuzzy.
function renderLocal() {
	if (!gridEl) return;

	// Purge de l'injection précédente (reconnexion, suppression, édition).
	var stale = document.querySelectorAll("[data-local-entry]");
	for (var s = 0; s < stale.length; s++) stale[s].remove();

	// Un filtre tag actif suspend l'injection : les tags locaux ne sont
	// pas filtrables, les afficher ici resterait trompeur.
	if (gridEl.getAttribute("data-tag-filtered") === "1") return;

	if (emptyEl) emptyEl.classList.toggle("hidden", entries.length > 0);

	for (var i = 0; i < entries.length; i++) {
		gridEl.appendChild(cardNode(entries[i]));
		if (tableRowsEl) tableRowsEl.appendChild(rowNode(entries[i]));
	}
	if (window.grzFuzzyRefresh) window.grzFuzzyRefresh();
}

// Carte locale : clone exact de kbCard (canon console) — heading
// ID · module · date, titre = 1re ligne de l'Erreur, extrait, pied avec
// tag principal, premier T-code et action « Consulter → ». Le module et
// les T-codes viennent du serveur (parse endpoint, mêmes règles Go que les
// cartes partagées : aucune dérive possible). data-local-entry porte la
// purge ; l'ancre étirée ouvre la fiche dans le tiroir, comme une carte
// serveur ouvre la sienne. Modifier/Supprimer vivent dans la fiche (comme
// côté serveur), pas sur la carte.
function cardNode(e) {
	var card = el("div", "relative bg-g-widget rounded-xl p-5 border border-g-border hover:border-g-cyan transition-colors");
	card.setAttribute("data-local-entry", "1");
	card.setAttribute("data-search", searchKey(e));

	var overlay = el("a", "absolute inset-0 rounded-xl");
	overlay.href = "#" + encodeURIComponent(e.id);
	overlay.setAttribute("data-local-detail", e.id);
	overlay.addEventListener("click", function (event) {
		event.preventDefault();
		openDetail(e);
	});
	card.appendChild(overlay);

	var header = el("div", "relative z-10 flex flex-wrap items-baseline gap-x-2 gap-y-0.5 justify-between");
	var meta = el("div", "relative z-10 flex flex-wrap items-baseline gap-x-2 gap-y-0.5");
	meta.appendChild(el("span", "font-mono text-xs font-semibold text-g-accent-text", e.id));
	if (e.parsed.module) {
		meta.appendChild(el("span", "text-g-subdue", "·"));
		meta.appendChild(el("span", "text-xs text-g-text", e.parsed.module));
	}
	header.appendChild(meta);
	var dateLabel = dateOf(e);
	if (dateLabel) header.appendChild(el("span", "tabular-nums font-mono text-[11px] text-g-text", dateLabel));
	card.appendChild(header);

	// Titre/extrait : même règle que cardTitleAndExcerpt (1re ligne de
	// l'Erreur, extrait = lignes suivantes bornées à 140).
	var titre = firstLine(e.parsed.erreur);
	if (titre) card.appendChild(el("p", "relative z-0 text-sm font-semibold text-g-hi line-clamp-2 mt-2", titre));
	var rest = afterFirstLine(e.parsed.erreur);
	if (rest) card.appendChild(el("p", "relative z-0 text-xs text-g-text line-clamp-2 mt-1", truncateTo(firstLine(rest), 140)));

	var footer = el("div", "relative z-10 flex flex-wrap items-center gap-2 mt-3");
	var tags = e.parsed.tags || [];
	if (tags.length) footer.appendChild(tagChip(tags[0]));
	var codes = e.parsed.tcodes || [];
	if (codes.length) footer.appendChild(el("span", "font-mono text-[11px] text-g-text", codes[0]));
	footer.appendChild(el("span", "ml-auto text-xs text-g-text", "Consulter →"));
	card.appendChild(footer);
	return card;
}

// Ligne locale pour la vue tableau : clone exact de kbTableRow — ID mono
// cliquable (ouvre la fiche dans le tiroir), module dérivé, titre tronqué,
// tag principal, premier T-code, date. Les dérivations viennent du parse
// endpoint, comme pour les cartes.
function rowNode(e) {
	var tr = el("tr", "h-9 hover:bg-g-hover transition-colors border-b border-g-border");
	tr.setAttribute("data-local-entry", "1");
	tr.setAttribute("data-search", searchKey(e));

	var idCell = el("td", "px-4 font-mono text-xs");
	var idLink = el("a", "text-g-accent-text hover:underline", e.id);
	idLink.href = "#" + encodeURIComponent(e.id);
	idLink.addEventListener("click", function (event) {
		event.preventDefault();
		openDetail(e);
	});
	idCell.appendChild(idLink);
	tr.appendChild(idCell);

	tr.appendChild(el("td", "px-4 text-xs text-g-text", e.parsed.module || "—"));

	var titleCell = el("td", "px-4 text-g-text");
	titleCell.appendChild(el("div", "max-w-2xl truncate", truncateTo(firstLine(e.parsed.erreur), 90)));
	tr.appendChild(titleCell);

	var tags = e.parsed.tags || [];
	tr.appendChild(el("td", "px-4 font-mono text-[11px] text-g-text", tags.length ? tags[0] : "—"));
	var codes = e.parsed.tcodes || [];
	tr.appendChild(el("td", "px-4 font-mono text-[11px] text-g-text", codes.length ? codes[0] : "—"));

	var dateCell = el("td", "px-4 text-xs text-g-text");
	var dateLabel = dateOf(e);
	if (dateLabel) dateCell.appendChild(el("span", "tabular-nums", dateLabel));
	tr.appendChild(dateCell);
	return tr;
}

// ---- fiche locale dans le tiroir latéral --------------------------------
// Capture PARESSEUSE : le tiroir peut être rendu après le script selon
// l'ordre d'assemblage de la page — on résout les éléments à chaque usage,
// jamais au chargement (cette fragilité d'ordre a tué l'ouverture des
// fiches locales une fois déjà).
var lastTrigger = null;

function drawerRoots() {
	return {
		drawer: document.getElementById("kb-drawer"),
		body: document.getElementById("kb-drawer-body"),
		overlay: document.getElementById("kb-drawer-overlay"),
		close: document.getElementById("kb-drawer-close")
	};
}

function setBusy(form, busy) {
	var buttons = form.querySelectorAll('button[type="submit"], button:not([type])');
	for (var i = 0; i < buttons.length; i++) buttons[i].disabled = busy;
}

// innerHTML n'exécute pas les <script> injectés : les boutons de copie du
// fragment sont câblés ici (repli simple, sans la bascule d'icône).
function bindCodeCopy(root) {
	var buttons = root.querySelectorAll(".md-code-copy");
	for (var i = 0; i < buttons.length; i++) {
		(function (button) {
			button.addEventListener("click", function () {
				var code = button.parentElement.querySelector("pre > code");
				if (!code || !navigator.clipboard) return;
				navigator.clipboard.writeText(code.textContent || "").catch(function () {});
			});
		})(buttons[i]);
	}
}

function openDrawer() {
	var roots = drawerRoots();
	if (!roots.drawer) return;
	roots.body.scrollTop = 0;
	roots.overlay.classList.add("open");
	roots.drawer.classList.add("open");
	roots.close.focus();
	// verrouille le défilement de la page derrière le tiroir
	document.body.style.overflow = "hidden";
}

function closeDrawer() {
	var roots = drawerRoots();
	if (!roots.drawer || !roots.drawer.classList.contains("open")) return;
	roots.overlay.classList.remove("open");
	roots.drawer.classList.remove("open");
	document.body.style.overflow = "";
	roots.body.innerHTML = "";
	if (lastTrigger) lastTrigger.focus();
}

(function () {
	// Fermeture livrée une fois pour toutes : l'audit des éléments est
	// refait au moment de l'usage, pas au chargement.
	document.addEventListener("click", function (event) {
		var roots = drawerRoots();
		if (!roots.overlay || !roots.drawer) return;
		if (event.target === roots.overlay || event.target === roots.close) closeDrawer();
	});
	document.addEventListener("keydown", function (event) {
		var roots = drawerRoots();
		if (event.key === "Escape" && roots.drawer && roots.drawer.classList.contains("open")) closeDrawer();
	});
})();

// addLocalSection : clone de noteSection — petit titre gras + corps, une
// section ne se rend que si son contenu est non vide (fiche adaptative).
function addLocalSection(host, titre, contenu, buildBody) {
	if (!contenu || !contenu.toString().replace(/\s+/g, "").length) return;
	var section = el("div", "space-y-1.5");
	section.appendChild(el("h2", "text-sm font-semibold text-g-hi", titre));
	section.appendChild(buildBody(contenu));
	host.appendChild(section);
}

// openDetail construit la fiche locale comme un clone de
// PageKBDetailPanel : bandeau ID · module · date, chips T-codes (/n),
// Erreur en trace brute mono (jamais rendue markdown), Notes/Résolution
// rendues par la pipeline goldmark du serveur (liens wiki résolus contre
// les IDs partagés), fiche adaptative, actions Modifier/Supprimer.
async function openDetail(e) {
	var roots = drawerRoots();
	if (!roots.drawer) return;
	var codes = e.parsed.tcodes || [];
	var response = await fetch("/kb/local/render", {
		method: "POST",
		headers: { "Content-Type": "application/json; charset=utf-8" },
		body: JSON.stringify({ notes: e.parsed.notes || "", resolution: e.parsed.resolution || "" })
	});
	if (!response.ok) {
		setMessage("Affichage impossible : le serveur de rendu est indisponible.");
		return;
	}
	var rendered = await response.json();

	lastTrigger = document.activeElement;
	var host = el("div", "px-6 py-5 space-y-5");

	// Bandeau : identifiant mono + module dérivé + date (canon panelHeader).
	var headRow = el("div", "flex items-start justify-between gap-3 flex-wrap");
	var meta = el("div", "flex items-baseline gap-2 flex-wrap");
	meta.appendChild(el("span", "font-mono text-lg font-semibold text-g-hi", e.id));
	if (e.parsed.module) {
		meta.appendChild(el("span", "text-g-subdue", "·"));
		meta.appendChild(el("span", "text-xs text-g-text", e.parsed.module));
	}
	headRow.appendChild(meta);
	var dateLabel = dateOf(e);
	if (dateLabel) headRow.appendChild(el("span", "tabular-nums font-mono text-[11px] text-g-text", dateLabel));
	host.appendChild(headRow);

	// Barre d'actions en tête de fiche, comme la fiche serveur (page
	// détail / panel) : Modifier + Supprimer visibles sans défiler.
	var actions = el("div", "flex justify-end gap-2");
	var edit = el("a", "mx-2 mb-1 px-3.5 py-1.5 rounded-lg bg-g-block text-g-text hover:text-g-hi hover:bg-g-hover text-xs transition-colors", "Modifier");
	edit.href = "/kb/new?local=1&id=" + encodeURIComponent(e.id);
	actions.appendChild(edit);
	var del = el("button", "px-3.5 py-1.5 rounded-lg bg-g-block text-g-negative-text hover:bg-g-hover hover:brightness-125 text-xs transition-colors", "Supprimer");
	del.type = "button";
	del.addEventListener("click", function () { removeEntry(e.id); });
	actions.appendChild(del);
	host.appendChild(actions);

	// Bandeau T-codes : chips mono notation /n, prêts pour la commande SAP.
	if (codes.length) {
		var banner = el("div", "flex flex-wrap items-center gap-1.5");
		for (var c = 0; c < codes.length; c++) {
			banner.appendChild(el("span", "px-2 py-0.5 rounded bg-g-code border border-g-border font-mono text-[11px] text-g-accent-text", "/n" + codes[c]));
		}
		host.appendChild(banner);
	}

	// Erreur : trace brute (mono, échappée), jamais du markdown rendu.
	addLocalSection(host, "Erreur", e.parsed.erreur, function (texte) {
		return el("pre", "font-mono text-xs leading-relaxed whitespace-pre-wrap break-words text-g-hi bg-g-code border border-g-border rounded-lg p-3", texte);
	});
	// Notes / Résolution : markdown sûr rendu par le serveur (.md).
	addLocalSection(host, "Notes", e.parsed.notes, function () {
		var node = el("div", "md");
		node.innerHTML = rendered.notesHTML || "";
		return node;
	});
	addLocalSection(host, "Résolution", e.parsed.resolution, function () {
		var node = el("div", "md");
		node.innerHTML = rendered.resolutionHTML || "";
		return node;
	});

	roots.body.innerHTML = "";
	roots.body.appendChild(host);
	bindCodeCopy(roots.body);
	openDrawer();
}

async function removeEntry(id) {
	if (!dir) return;
	if (!window.confirm("Supprimer le fichier local " + id + ".md ?")) return;
	try {
		await dir.removeEntry(id + ".md");
		closeDrawer();
		refreshIndex();
	} catch (err) {
		setMessage("Suppression impossible : " + ((err && err.message) || err));
	}
}

// ---- enregistrement local depuis les formulaires existants ---------------
// Bascule cochée immédiatement : l'URL ?local=1 n'est posée que par le
// bouton Modifier de la fiche locale — l'article est local par
// construction. La case (et le bloc) deviennent visibles/cochés
// SYNCHRONIQUEMENT, sans attendre le réseau ni IndexedDB.
var editingLocal = !!(toggle && new URLSearchParams(window.location.search).get("local") === "1");
if (editingLocal) {
	toggle.checked = true;
	revealBlock();
}

var validLocalID = function (id) { return /^[A-Za-z0-9_-]+$/.test(id); };

// Pré-remplissage du mode édition locale (?local=1&id=...) : lecture du
// fichier, analyse par le serveur, remplissage des champs présents.
if (editingLocal) {
	(async function () {
		await discovery;
		if (!dir) {
			setMessage("Dossier non connecté : l'enregistrement restera en attente (connectez le dossier puis rechargez).");
			return;
		}
		var id = new URLSearchParams(window.location.search).get("id");
		if (!id || !validLocalID(id)) return;
		var form = document.querySelector("form[data-local-save]");
		if (!form) return;
		var handle;
		try {
			handle = await dir.getFileHandle(id + ".md");
		} catch (err) {
			return; // identifiant libre : ne pas écraser la saisie du formulaire
		}
		var raw = await (await handle.getFile()).text();
		var response = await fetch("/kb/local/parse", {
			method: "POST",
			headers: { "Content-Type": "text/markdown; charset=utf-8" },
			body: raw
		});
		if (!response.ok) return;
		var parsed = await response.json();
		var values = { "erreur": parsed.erreur, "notes": parsed.notes, "resolution": parsed.resolution, "tags": (parsed.tags || []).join(", ") };
		for (var name in values) {
			var input = form.elements[name];
			if (input && values[name]) input.value = values[name];
		}
	})();
}

// Intercept en capture : stopPropagation empêche htmx (listener posé sur
// le form) et le POST natif d'aboutir quand l'enregistrement est local.
document.addEventListener("submit", function (event) {
	var form = event.target;
	if (!form.matches || !form.matches("form[data-local-save]")) return;
	if (!toggle || !toggle.checked) return;
	event.preventDefault();
	event.stopPropagation();
	saveLocal(form);
}, true);

async function saveLocal(form) {
	setBusy(form, true);
	try {
		if (!dir) {
			showFormMessage("Dossier local non connecté : enregistrement en attente (bouton « Connecter »).");
			return;
		}
		var idInput = form.elements["id"];
		var id = idInput ? idInput.value.trim() : "";
		var body = new URLSearchParams();
		var fields = form.elements;
		for (var i = 0; i < fields.length; i++) {
			if (fields[i].name) body.set(fields[i].name, fields[i].value);
		}
		var response = await fetch("/kb/local/serialize", {
			method: "POST",
			headers: { "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8" },
			body: body.toString()
		});
		var content = await response.text();
		if (!response.ok) {
			showFormMessage(content || "Enregistrement local refusé.");
			return;
		}
		var handle = await dir.getFileHandle(id + ".md", { create: true });
		var writable = await handle.createWritable();
		await writable.write(content);
		await writable.close();
		showFormMessage("KB " + id + " enregistrée dans votre dossier local « " + dir.name + " ».", id !== "");
	} catch (err) {
		showFormMessage("Enregistrement local impossible : " + ((err && err.message) || err));
	} finally {
		setBusy(form, false);
	}
}

function showFormMessage(text, withHomeLink) {
	var zone = document.getElementById("reponse-serveur");
	if (!zone) return;
	zone.textContent = text;
	if (withHomeLink) {
		var link = document.createElement("a");
		link.href = "/";
		link.className = "text-g-accent-text underline";
		link.textContent = "Voir mes articles";
		zone.appendChild(document.createTextNode(" "));
		zone.appendChild(link);
	}
}
})();`

// localKBScriptNode retourne l'élément <script> du comportement local.
func localKBScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(localKBScript))
}

// LocalSaveToggle retourne la bascule d'enregistrement local pour les
// formulaires de KB (création complète et création libre). Cachée par
// défaut : le script ne la révèle qu'en présence de File System Access.
// L'attribut data-local-block sert d'ancre de révélation, le marqueur
// data-local-save guide l'interception du submit.
func LocalSaveToggle() gomponents.Node {
	return gomponents.El("div", gomponents.Attr("data-local-block"),
		Class("hidden space-y-1"),
		gomponents.El("label",
			Class("flex items-center gap-2 text-g-text text-xs"),
			gomponents.El("input",
				gomponents.Attr("type", "checkbox"),
				ID("kb-local-save"),
				Class("mt-0"),
			),
			gomponents.Text("Enregistrer dans mon dossier local (ce poste uniquement)"),
		),
		Button(
			Type("button"),
			gomponents.Attr("data-local-connect", "1"),
			Class("px-3 py-1.5 rounded-md border border-g-border bg-g-widget text-g-text hover:text-g-hi hover:border-g-cyan transition-colors text-xs font-mono"),
			gomponents.Text("Connecter mon dossier"),
		),
		Span(ID("local-kb-status"), Class("text-g-text text-xs")),
		gomponents.El("script", gomponents.Raw(localKBScript)),
	)
}

// PageLocalBar retourne la barre « Articles locaux » de l'accueil :
// libellé, connexion du dossier, état. La barre est créée cachée (hidden)
// : sans JavaScript ou sans support navigateur, l'accueil est inchangé ;
// les cartes locales, elles, s'invitent directement dans la grille
// partagée #kb-cards via le script.
func PageLocalBar() gomponents.Node {
	return gomponents.El("div",
		ID("local-kb-bar"),
		gomponents.Attr("data-local-block"),
		Class("hidden max-w-5xl mx-auto px-6 pb-2"),
		Div(Class("flex flex-wrap items-center gap-x-3 gap-y-1"),
			Button(
				Type("button"),
				gomponents.Attr("data-local-connect", "1"),
				Class("px-3 py-1.5 rounded-md border border-g-border bg-g-widget text-g-text hover:text-g-hi hover:border-g-cyan transition-colors text-xs font-mono"),
				gomponents.Text("Connecter mon dossier"),
			),
			Span(ID("local-kb-status"), Class("text-g-text text-xs")),
		),
		localKBScriptNode(),
	)
}
