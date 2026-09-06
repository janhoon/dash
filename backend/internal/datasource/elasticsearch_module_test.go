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

	acees "github.com/aceobservability/ace-datasource-elasticsearch"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestElasticsearchModule_QueryAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawSearch, sawHealth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "elastic" || password != "secret" {
			t.Errorf("expected basic auth elastic/secret, got %s/%s ok=%v", username, password, ok)
		}
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/_search"):
			sawSearch = true
			if r.URL.Path != "/logs-*/_search" {
				t.Errorf("search path=%s, want /logs-*/_search", r.URL.Path)
			}
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			query, ok := body["query"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected query object, got %+v", body)
			}
			if _, hasBool := query["bool"]; !hasBool {
				t.Fatalf("expected bool query with time filter, got %+v", query)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"hits": {
					"hits": [
						{
							"_index": "logs-2026.02.22",
							"_id": "log-1",
							"_source": {
								"@timestamp": "2026-02-22T15:00:00Z",
								"message": "request failed",
								"level": "ERROR",
								"service.name": "api"
							}
						}
					]
				}
			}`))
		case r.URL.Path == "/_cluster/health":
			sawHealth = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"green"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{
		Type:     models.DataSourceElasticsearch,
		URL:      srv.URL,
		AuthType: "basic",
		AuthConfig: []byte(`{
			"username":"elastic",
			"password":"secret",
			"index":"logs-*"
		}`),
	}

	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, ok := client.(*acees.Client); !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	qws, ok := client.(SignalQueryClient)
	if !ok {
		t.Fatalf("elasticsearch client must implement SignalQueryClient, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Unix(1_700_000_000, 0)
	end := start.Add(15 * time.Minute)
	result, err := qws.QueryWithSignal(ctx, "error", "logs", start, end, time.Minute, 200)
	if err != nil {
		t.Fatalf("QueryWithSignal: %v", err)
	}
	if result.ResultType != "logs" {
		t.Fatalf("ResultType=%q, want logs", result.ResultType)
	}
	if result.Data == nil || len(result.Data.Logs) != 1 {
		t.Fatalf("expected 1 log, got %+v", result.Data)
	}
	entry := result.Data.Logs[0]
	if entry.Line != "request failed" {
		t.Fatalf("line=%q", entry.Line)
	}
	if entry.Level != "error" {
		t.Fatalf("level=%q", entry.Level)
	}
	if !sawSearch {
		t.Fatal("expected registry client to hit fixture /_search")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawHealth {
		t.Fatal("expected TestConnection to hit fixture /_cluster/health")
	}
}

func TestElasticsearch_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{
		Type: models.DataSourceElasticsearch,
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
