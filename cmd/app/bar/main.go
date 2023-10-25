package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider(ctx context.Context) (func(context.Context) error, error) {

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName("bar-service"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", "simple-prod-collector", 4317),
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func main() {
	log.Printf("Waiting for connection...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initProvider(ctx)
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
		BaseContext:  func(net.Listener) context.Context { return ctx },
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
		bag, _ := baggage.Parse("username=donuts")
		ctx := baggage.ContextWithBaggage(context.Background(), bag) // TODO try with request context
		// Attributes represent additional key-value descriptors that can be bound
		// to a metric observer or recorder.
		commonAttrs := []attribute.KeyValue{
			attribute.String("attrA", "chocolate"),
			attribute.String("attrB", "raspberry"),
			attribute.String("attrC", "vanilla"),
		}

		// work begins
		ctx, span := tracer.Start(
			ctx,
			"CollectorExporter-Example-BAR",
			trace.WithAttributes(commonAttrs...))
		defer span.End()

		req, err := http.NewRequest(http.MethodGet, "http://foo:8080/foo", nil)
		req.WithContext(ctx)
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
	})

	return mux
}
