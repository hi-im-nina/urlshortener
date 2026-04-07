package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const baseURL = "http://localhost:" + port

	store := NewStore()
	h := &Handlers{
		store:   store,
		baseURL: baseURL,
	}

	// Go's standard library uses a "mux" (multiplexer) to route requests to
	// the right handler based on the URL path.
	//
	// A more full-featured router (like gorilla/mux or chi) gives you things
	// like path parameters (/users/{id}) and middleware — but for a beginner
	// project the standard library is perfect and has no dependencies.
	mux := http.NewServeMux()

	// POST /shorten  — create a short URL
	mux.HandleFunc("/shorten", h.Shorten)

	// GET /stats/{code}  — must be registered before "/" or it gets swallowed
	mux.HandleFunc("/stats/", h.Stats)

	// GET /{code}  — catch-all: redirect any other path
	mux.HandleFunc("/", h.Redirect)

	fmt.Printf("URL shortener running at %s\n", baseURL)
	fmt.Println("  POST /shorten           — create a short URL")
	fmt.Println("  GET  /{code}            — redirect to original URL")
	fmt.Println("  GET  /stats/{code}      — view click stats")
	fmt.Println()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
