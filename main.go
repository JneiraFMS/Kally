package main

import (
	"log"
	"net/http"
	"path/filepath"

	"grz/internal/adapters/outbound/persistence"
	"grz/internal/application/usecases"
	"grz/web"
)

const (
	kbDataDir    = "data/kb"
	mediaDataDir = "data/media"
)

func main() {
	// Root de composition : les dépendances concrètes sont créées ici
	// puis injectées dans les adapters inbound.
	repo := persistence.NewMDRepository(kbDataDir)
	kbApp := usecases.NewKBService(repo)
	kbHandler := web.NewKBHandler(kbApp)

	// Le journal partage le répertoire de stockage mais suit ses propres
	// conventions de fichiers (journal-<date>.md, markdown libre).
	journalRepo := persistence.NewMDJournalRepository(kbDataDir)
	journalApp := usecases.NewJournalService(journalRepo)
	journalHandler := web.NewJournalHandler(journalApp)

	// Images collées : stockage content-addressed, servi par un adapter HTTP
	// dédié (POST /media, GET /media/{name}).
	mediaRepo := persistence.NewLocalMediaRepository(mediaDataDir)
	mediaApp := usecases.NewMediaService(mediaRepo)
	mediaHandler := web.NewMediaHandler(mediaApp)

	mux := http.NewServeMux()

	// Feuille de styles locale (Tailwind généré), embarquée dans le binaire
	mux.HandleFunc("GET /assets/tailwind.css", web.ServeStylesheet)

	// Images collées dans les notes (markdown)
	mux.HandleFunc("POST /media", mediaHandler.Upload)
	mux.HandleFunc("GET /media/{name}", mediaHandler.Serve)

	// Page d'accueil : cartes des KB existantes
	mux.HandleFunc("GET /{$}", kbHandler.Home)

	// Journal (inspiré Logseq) : une note libre datée, plus de routes KB
	mux.HandleFunc("GET /journal", journalHandler.Page)
	mux.HandleFunc("GET /journal/new", journalHandler.ShowNewForm)
	mux.HandleFunc("POST /journal", journalHandler.Create)
	mux.HandleFunc("GET /journal/{date}", journalHandler.ShowEntry)
	mux.HandleFunc("GET /journal/{date}/edit", journalHandler.ShowEditForm)
	mux.HandleFunc("POST /journal/edit", journalHandler.UpdateEntry)
	mux.HandleFunc("POST /journal/delete", journalHandler.Delete)

	// Pages KB
	mux.HandleFunc("GET /kb", kbHandler.List)
	mux.HandleFunc("GET /kb/graph", kbHandler.ShowGraph)
	mux.HandleFunc("GET /kb/new", kbHandler.ShowNewForm)
	mux.HandleFunc("POST /kb", kbHandler.Create)
	mux.HandleFunc("GET /kb/detail", kbHandler.Detail)
	mux.HandleFunc("GET /kb/edit", kbHandler.ShowEditForm)
	mux.HandleFunc("POST /kb/edit", kbHandler.UpdateForm)
	mux.HandleFunc("POST /kb/delete", kbHandler.Delete)

	// Affiche où iront les fichiers .md (relatif au répertoire de lancement)
	absDir, err := filepath.Abs(kbDataDir)
	if err != nil {
		absDir = kbDataDir
	}
	log.Printf("Stockage des KB : %s", absDir)

	// Même logique pour les images collées
	absMediaDir, err := filepath.Abs(mediaDataDir)
	if err != nil {
		absMediaDir = mediaDataDir
	}
	log.Printf("Stockage des images : %s", absMediaDir)

	// Lancement du serveur : une erreur est fatale et affichée
	// (ex: port 8080 déjà occupé par une ancienne instance)
	log.Println("Serveur démarré sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
