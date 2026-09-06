package datasource

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	acevl "github.com/aceobservability/ace-datasource-victorialogs"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestVictoriaLogsModule_QueryAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawQuery, sawHealth, sawFieldNames, sawFieldValues bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/select/logsql/query":
			sawQuery = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"_msg":"boom","_time":"2026-02-08T12:00:00Z","service":"api","level":"error"}`+"\n")
		case "/health":
			sawHealth = true
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "OK")
		case "/select/logsql/field_names":
			sawFieldNames = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"values":[{"value":"level"},{"value":"service"}]}`)
		case "/select/logsql/field_values":
			sawFieldValues = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"values":[{"value":"api"},{"value":"web"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{
		Type: models.DataSourceVictoriaLogs,
		URL:  srv.URL,
	}

	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	vlogs, ok := client.(*acevl.Client)
	if !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Unix(1600000000, 0).Add(-time.Hour)
	end := time.Unix(1600000000, 0)
	result, err := client.Query(ctx, "*", start, end, time.Minute, 10)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result.Status != "success" {
		t.Fatalf("Query status=%q error=%q", result.Status, result.Error)
	}
	if result.Data == nil || len(result.Data.Logs) != 1 || result.Data.Logs[0].Line != "boom" {
		t.Fatalf("unexpected query result %+v", result.Data)
	}
	if !sawQuery {
		t.Fatal("expected registry client to hit fixture /select/logsql/query")
	}

	labels, err := vlogs.Labels(ctx)
	if err != nil {
		t.Fatalf("Labels: %v", err)
	}
	if len(labels) != 2 || labels[0] != "level" || labels[1] != "service" {
		t.Fatalf("Labels=%v", labels)
	}
	if !sawFieldNames {
		t.Fatal("expected Labels to hit fixture /select/logsql/field_names")
	}

	values, err := vlogs.LabelValues(ctx, "service")
	if err != nil {
		t.Fatalf("LabelValues: %v", err)
	}
	if len(values) != 2 || values[0] != "api" {
		t.Fatalf("LabelValues=%v", values)
	}
	if !sawFieldValues {
		t.Fatal("expected LabelValues to hit fixture /select/logsql/field_values")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawHealth {
		t.Fatal("expected TestConnection to hit fixture /health")
	}
}

func TestVictoriaLogs_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{
		Type: models.DataSourceVictoriaLogs,
		URL:  "http://169.254.169.254/",
	}
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

func TestVictoriaLogsModule_StreamAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawTail bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/select/logsql/tail" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("method=%s, want POST", r.Method)
		}
		sawTail = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"_msg":"hello","_time":"2026-02-08T12:00:00Z","service":"api"}`+"\n")
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(models.DataSource{
		Type: models.DataSourceVictoriaLogs,
		URL:  srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	streamer, ok := client.(StreamClient)
	if !ok {
		t.Fatalf("registry client must implement StreamClient, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var gotLine string
	err = streamer.Stream(ctx, `service:api`, time.Time{}, 1, func(entry LogEntry) error {
		gotLine = entry.Line
		cancel()
		return nil
	})
	if err != nil && ctx.Err() == nil {
		t.Fatalf("Stream: %v", err)
	}
	if !sawTail {
		t.Fatal("expected registry client to POST fixture /select/logsql/tail")
	}
	if gotLine != "hello" {
		t.Fatalf("line=%q, want hello", gotLine)
	}
}
