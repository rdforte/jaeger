package main

import (
	"Tracing/internal/server"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/baggage"
	"html"
	"io"
	"net/http"
	"os"
)

// TODO - all otel/tracing logic should be part of middleware and hidden away from main handler logic

func main() {
	if err := server.RunServer("bar-service", newHTTPHandler()); err != nil {
		fmt.Printf("bar service error %v", err)
		os.Exit(1)
	}
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	barHandler := func(w http.ResponseWriter, r *http.Request) {
		m, err := baggage.NewMember("clientID", "donuts")
		if err != nil {
			fmt.Println(err)
		}
		bag, err := baggage.New(m)
		ctx := baggage.ContextWithBaggage(r.Context(), bag)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://foo:8080/foo", nil)
		if err != nil {
			fmt.Println(err)
		}
		client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
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

	}
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(barHandler), "bar-handler")

	mux.Handle("/bar", otelHandler)

	return mux
}
