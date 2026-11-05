package pages

import "maragu.dev/gomponents"

// imagePasteScript câble le collage et le glisser-déposer d'images sur les
// textareas markdown : envoi binaire vers POST /media, puis insertion du
// markdown ![image](/media/<hash>.<ext>) à la position du curseur.
//
// Aucun état n'est conservé côté serveur : le script insère un marqueur
// unique pendant l'envoi, le remplace par le lien en cas de succès et le
// retire en cas d'échec. Les boutons du formulaire sont désactivés pendant
// l'upload pour éviter d'enregistrer avant la fin de l'insertion.
const imagePasteScript = `/* grz-image-paste */
(function () {
"use strict";

var ENDPOINT = "/media";
var MAX_SIZE = 5 * 1024 * 1024;
var counter = 0;

function insertAt(textarea, text) {
	var start = textarea.selectionStart;
	var end = textarea.selectionEnd;
	var value = textarea.value;
	textarea.value = value.slice(0, start) + text + value.slice(end);
	textarea.selectionStart = textarea.selectionEnd = start + text.length;
	textarea.dispatchEvent(new Event("input", { bubbles: true }));
}

function setBusy(textarea, busy) {
	var form = textarea.form;
	if (!form) return;
	var buttons = form.querySelectorAll('button[type="submit"], button:not([type])');
	for (var i = 0; i < buttons.length; i++) {
		buttons[i].disabled = busy;
	}
}

function upload(textarea, file) {
	if (!file || file.type.indexOf("image/") !== 0) return Promise.resolve();
	if (file.size > MAX_SIZE) {
		window.alert("Image trop volumineuse (5 Mo maximum)");
		return Promise.resolve();
	}
	counter += 1;
	var marker = "![envoi de l'image " + counter + "…]()";
	insertAt(textarea, marker);
	setBusy(textarea, true);
	return fetch(ENDPOINT, {
		method: "POST",
		headers: { "Content-Type": file.type },
		body: file
	}).then(function (response) {
		if (!response.ok) {
			return response.json().catch(function () { return {}; }).then(function (data) {
				throw new Error((data && data.error) || "Envoi de l'image impossible");
			});
		}
		return response.json();
	}).then(function (data) {
		textarea.value = textarea.value.replace(marker, "![image](" + data.url + ")");
		textarea.dispatchEvent(new Event("input", { bubbles: true }));
	}).catch(function (error) {
		textarea.value = textarea.value.replace(marker, "");
		textarea.dispatchEvent(new Event("input", { bubbles: true }));
		window.alert(error.message || "Envoi de l'image impossible");
	}).then(function () {
		setBusy(textarea, false);
	});
}

function uploadAll(textarea, files) {
	var images = [];
	for (var i = 0; i < files.length; i++) {
		if (files[i] && files[i].type && files[i].type.indexOf("image/") === 0) {
			images.push(files[i]);
		}
	}
	var chain = Promise.resolve();
	for (var j = 0; j < images.length; j++) {
		(function (file) {
			chain = chain.then(function () { return upload(textarea, file); });
		})(images[j]);
	}
}

function onPaste(event) {
	var data = event.clipboardData;
	if (!data || !data.items) return;
	var files = [];
	for (var i = 0; i < data.items.length; i++) {
		if (data.items[i].kind === "file") {
			var file = data.items[i].getAsFile();
			if (file && file.type.indexOf("image/") === 0) files.push(file);
		}
	}
	if (!files.length) return;
	event.preventDefault();
	uploadAll(this, files);
}

function hasFiles(dataTransfer) {
	if (!dataTransfer || !dataTransfer.types) return false;
	var types = dataTransfer.types;
	for (var i = 0; i < types.length; i++) {
		if (types[i] === "Files") return true;
	}
	return false;
}

function highlight(textarea, on) {
	textarea.style.outline = on ? "2px dashed var(--g-sep)" : "";
	textarea.style.outlineOffset = on ? "-4px" : "";
}

function onDragOver(event) {
	if (!hasFiles(event.dataTransfer)) return;
	event.preventDefault();
	event.dataTransfer.dropEffect = "copy";
	highlight(this, true);
}

function onDragLeave() {
	highlight(this, false);
}

function onDrop(event) {
	highlight(this, false);
	if (!event.dataTransfer || !event.dataTransfer.files || !event.dataTransfer.files.length) return;
	event.preventDefault();
	uploadAll(this, event.dataTransfer.files);
}

var areas = document.querySelectorAll("textarea[data-image-paste]");
for (var i = 0; i < areas.length; i++) {
	areas[i].addEventListener("paste", onPaste);
	areas[i].addEventListener("dragover", onDragOver);
	areas[i].addEventListener("dragleave", onDragLeave);
	areas[i].addEventListener("drop", onDrop);
}
})();`

// imagePasteScriptNode retourne le script à inclure dans les pages de
// formulaire (jamais dans les pages de lecture).
func imagePasteScriptNode() gomponents.Node {
	return gomponents.El("script", gomponents.Raw(imagePasteScript))
}
