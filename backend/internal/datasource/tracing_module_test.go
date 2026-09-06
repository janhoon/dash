package datasource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	acetempo "github.com/aceobservability/ace-datasource-tempo"
	acevt "github.com/aceobservability/ace-datasource-victoriatraces"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestTempoModule_GetTraceAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawReady, sawTrace bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/ready":
			sawReady = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		case strings.HasPrefix(r.URL.Path, "/api/traces/"):
			sawTrace = true
			_, _ = w.Write([]byte(`{"data":[{"traceID":"trace-123","spans":[{"traceID":"trace-123","spanID":"root","operationName":"GET /","references":[],"startTime":1700000000000000,"duration":1000,"tags":[],"processID":"p1"}],"processes":{"p1":{"serviceName":"frontend"}}}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{Type: models.DataSourceTempo, URL: srv.URL}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, ok := client.(*acetempo.Client); !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tracingClient, err := NewTracingClient(ds)
	if err != nil {
		t.Fatalf("NewTracingClient: %v", err)
	}
	trace, err := tracingClient.GetTrace(ctx, "trace-123")
	if err != nil {
		t.Fatalf("GetTrace: %v", err)
	}
	if trace.TraceID != "trace-123" {
		t.Fatalf("trace id=%q", trace.TraceID)
	}
	if !sawTrace {
		t.Fatal("expected registry client to hit fixture /api/traces/")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawReady {
		t.Fatal("expected TestConnection to hit fixture /ready")
	}
}

func TestVictoriaTracesModule_ServicesAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawHealth, sawServices bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			sawHealth = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case "/select/jaeger/api/services":
			sawServices = true
			_, _ = w.Write([]byte(`{"data":["frontend","worker"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{Type: models.DataSourceVictoriaTraces, URL: srv.URL}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, ok := client.(*acevt.Client); !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tracingClient, err := NewTracingClient(ds)
	if err != nil {
		t.Fatalf("NewTracingClient: %v", err)
	}
	services, err := tracingClient.Services(ctx)
	if err != nil {
		t.Fatalf("Services: %v", err)
	}
	if len(services) != 2 || services[0] != "frontend" {
		t.Fatalf("Services=%v", services)
	}
	if !sawServices {
		t.Fatal("expected registry client to hit fixture /select/jaeger/api/services")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawHealth {
		t.Fatal("expected TestConnection to hit fixture /health")
	}
}

func TestTempo_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{Type: models.DataSourceTempo, URL: "http://169.254.169.254/"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := TestConnection(ctx, ds)
	if err == nil {
		t.Fatal("expected TestConnection to reject cloud metadata URL")
	}
	if !strings.Contains(err.Error(), "datasource url rejected") {
		t.Fatalf("error %q, want datasource url rejected", err)
	}
}

func TestVictoriaTraces_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{Type: models.DataSourceVictoriaTraces, URL: "http://169.254.169.254/"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := TestConnection(ctx, ds)
	if err == nil {
		t.Fatal("expected TestConnection to reject cloud metadata URL")
	}
	if !strings.Contains(err.Error(), "datasource url rejected") {
		t.Fatalf("error %q, want datasource url rejected", err)
	}
}
