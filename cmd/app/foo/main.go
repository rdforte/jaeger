package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"
)

func main() {
	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}

	log.Fatal(srv.ListenAndServe())
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Context())
		fmt.Println("Hit")
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	return mux
}
