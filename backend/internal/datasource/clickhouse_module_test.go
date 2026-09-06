package datasource

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	acech "github.com/aceobservability/ace-datasource-clickhouse"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestClickHouseModule_QueryAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawQuery, sawPing bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			sawQuery = true
			if r.URL.Query().Get("database") != "analytics" {
				t.Errorf("database=%q, want analytics", r.URL.Query().Get("database"))
			}
			username, password, ok := r.BasicAuth()
			if !ok || username != "alice" || password != "secret" {
				t.Errorf("basic auth=%v %s/%s, want alice/secret", ok, username, password)
			}
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			body := string(payload)
			if !strings.Contains(body, "FORMAT JSON") {
				t.Errorf("body missing FORMAT JSON: %q", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"timestamp":"2026-02-18T10:00:00Z","message":"ok","level":"info"}]}`))
		case r.URL.Path == "/ping":
			sawPing = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ok."))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{
		Type:     models.DataSourceClickHouse,
		URL:      srv.URL,
		AuthType: "basic",
		AuthConfig: json.RawMessage(`{
			"username":"alice",
			"password":"secret",
			"database":"analytics"
		}`),
	}

	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ch, ok := client.(*acech.Client)
	if !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Unix(1700000000, 0)
	end := start.Add(5 * time.Minute)
	result, err := ch.QueryWithSignal(ctx, "SELECT {start} AS start", "logs", start, end, 15*time.Second, 0)
	if err != nil {
		t.Fatalf("QueryWithSignal: %v", err)
	}
	if result.ResultType != "logs" {
		t.Fatalf("ResultType=%q, want logs", result.ResultType)
	}
	if result.Data == nil || len(result.Data.Logs) != 1 {
		t.Fatalf("unexpected query result %+v", result.Data)
	}
	if !sawQuery {
		t.Fatal("expected registry client to POST query to fixture")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawPing {
		t.Fatal("expected TestConnection to hit fixture /ping")
	}
}

func TestClickHouse_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{
		Type: models.DataSourceClickHouse,
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
