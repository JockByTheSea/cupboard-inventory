package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	initTemplates()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/pantry/add", addPantryHandler)
	mux.HandleFunc("/pantry/edit", editPantryHandler)
	mux.HandleFunc("/pantry/delete", deletePantryHandler)
	mux.HandleFunc("/freezer/add", addFreezerHandler)
	mux.HandleFunc("/freezer/edit", editFreezerHandler)
	mux.HandleFunc("/freezer/delete", deleteFreezerHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
