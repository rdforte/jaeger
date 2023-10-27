package main

import (
	"Tracing/internal/tracing"
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func main() {
	log.Printf("Waiting for connection...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := tracing.InitTracing(ctx, "foo-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	tracer = otel.Tracer("foo-tracer")
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

	fooHandler := func(w http.ResponseWriter, r *http.Request) {
		uk := attribute.Key("clientID")
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		bag := baggage.FromContext(ctx)
		span.AddEvent("clientID", trace.WithAttributes(uk.String(bag.Member("clientID").Value())))
		span.SetName("foo-handler")
		fmt.Fprintf(w, "Hello got your baggage: [ %v ]. FROM: %q", bag, html.EscapeString(r.URL.Path))
	}
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(fooHandler), "foo-handler")

	mux.Handle("/foo", otelHandler)

	return mux
}
