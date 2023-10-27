package server

import (
	"Tracing/internal/tracing"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func RunServer(serviceName string, handler http.Handler) {
	log.Printf("Waiting for connection...")
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shutdown, err := tracing.InitTracing(ctx, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	// Start HTTP server.
	srv := &http.Server{
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		Addr:         ":8080",
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}
