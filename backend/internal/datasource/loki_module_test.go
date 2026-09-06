package datasource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	aceloki "github.com/aceobservability/ace-datasource-loki"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestLokiModule_QueryAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawQueryRange, sawReady, sawLabels, sawLabelValues atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/loki/api/v1/query_range"):
			sawQueryRange.Store(true)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"streams","result":[{"stream":{"job":"ace"},"values":[["1700000000000000000","hello from loki"]]}]}}`))
		case r.URL.Path == "/ready":
			sawReady.Store(true)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		case strings.HasSuffix(r.URL.Path, "/loki/api/v1/labels") && !strings.Contains(r.URL.Path, "/values"):
			sawLabels.Store(true)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":["job","level"]}`))
		case strings.Contains(r.URL.Path, "/loki/api/v1/label/") && strings.HasSuffix(r.URL.Path, "/values"):
			sawLabelValues.Store(true)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":["ace","api"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{
		Type: models.DataSourceLoki,
		URL:  srv.URL,
	}

	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	loki, ok := client.(*aceloki.Client)
	if !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Unix(1700000000, 0).Add(-time.Hour)
	end := time.Unix(1700000000, 0)
	result, err := client.Query(ctx, `{job="ace"}`, start, end, time.Minute, 10)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result.Status != "success" {
		t.Fatalf("Query status=%q error=%q", result.Status, result.Error)
	}
	if result.Data == nil || len(result.Data.Logs) != 1 || result.Data.Logs[0].Line != "hello from loki" {
		t.Fatalf("unexpected query result %+v", result.Data)
	}
	if !sawQueryRange.Load() {
		t.Fatal("expected registry client to hit fixture /loki/api/v1/query_range")
	}

	labels, err := loki.Labels(ctx)
	if err != nil {
		t.Fatalf("Labels: %v", err)
	}
	if len(labels) != 2 || labels[0] != "job" || labels[1] != "level" {
		t.Fatalf("Labels=%v", labels)
	}
	if !sawLabels.Load() {
		t.Fatal("expected Labels to hit fixture /loki/api/v1/labels")
	}

	values, err := loki.LabelValues(ctx, "job")
	if err != nil {
		t.Fatalf("LabelValues: %v", err)
	}
	if len(values) != 2 || values[0] != "ace" {
		t.Fatalf("LabelValues=%v", values)
	}
	if !sawLabelValues.Load() {
		t.Fatal("expected LabelValues to hit fixture /loki/api/v1/label/job/values")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawReady.Load() {
		t.Fatal("expected TestConnection to hit fixture /ready")
	}
}

func TestLoki_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{
		Type: models.DataSourceLoki,
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
