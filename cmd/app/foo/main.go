package main

import (
	"Tracing/internal/server"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"html"
	"net/http"
	"os"
)

func main() {
	if err := server.RunServer("foo-service", newHTTPHandler()); err != nil {
		fmt.Printf("foo service error %v", err)
		os.Exit(1)
	}
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
