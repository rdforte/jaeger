package main

import (
	"Tracing/internal/tracing"
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func main() {
	log.Printf("Waiting for connection...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := tracing.InitTracing(ctx, "bar-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	tracer = otel.Tracer("bar-tracer")

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
