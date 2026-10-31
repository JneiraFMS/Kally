package main

import (
	"fmt"
	"grz/html/pages"
	"net/http"
)

func main() {
	// 1. Route pour afficher la page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// On passe r.URL.Path pour que la navbar sache quelle page est active !
		_ = pages.PageAccueil(r.URL.Path).Render(w)
	})

	// 2. Route pour traiter le formulaire soumis
	http.HandleFunc("/soumettre", func(w http.ResponseWriter, r *http.Request) {
		// Très important : analyse les données du formulaire
		_ = r.ParseForm()

		// Récupération de la valeur du champ "prenom"
		prenom := r.FormValue("prenom")

		// Petit traitement de sécurité / vérification
		if prenom == "" {
			prenom = "Anonyme"
		}

		fmt.Printf("Formulaire reçu pour : %s\n", prenom)

		// On renvoie le HTML qui va remplacer le contenu de #reponse-serveur
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fmt.Appendf(nil, "<span class='text-green-600'>Bonjour %s, formulaire bien reçu par le serveur ! 🚀</span>", prenom))
	})

	// Lancement du serveur
	http.ListenAndServe(":8080", nil)
}
