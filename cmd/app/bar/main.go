package main

import (
	"fmt"
	"html"
	"io"
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
	mux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://foo:8080/foo", nil)
		if err != nil {
			fmt.Println(err)
		}
		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------------")
		fmt.Println(string(body))
		fmt.Println("-------------------------------------")
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	return mux
}
