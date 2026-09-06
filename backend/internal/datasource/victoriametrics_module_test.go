package datasource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	acevm "github.com/aceobservability/ace-datasource-victoriametrics"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestVictoriaMetricsModule_QueryAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawQueryRange, sawHealth, sawLabels, sawMetricNames, sawLabelValues bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/api/v1/query_range"):
			sawQueryRange = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","job":"vm"},"values":[[1600000000,"1"]]}]}}`))
		case r.URL.Path == "/health":
			sawHealth = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case r.URL.Path == "/api/v1/labels":
			sawLabels = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":["__name__","job"]}`))
		case r.URL.Path == "/api/v1/label/__name__/values":
			sawMetricNames = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":["up","vm_rows"]}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/label/") && strings.HasSuffix(r.URL.Path, "/values"):
			sawLabelValues = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":["vm","node"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{
		Type: models.DataSourceVictoriaMetrics,
		URL:  srv.URL,
	}

	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	vm, ok := client.(*acevm.Client)
	if !ok {
		t.Fatalf("registry must return module client, got %T", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Unix(1600000000, 0).Add(-time.Hour)
	end := time.Unix(1600000000, 0)
	result, err := client.Query(ctx, "up", start, end, time.Minute, 0)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result.Status != "success" {
		t.Fatalf("Query status=%q error=%q", result.Status, result.Error)
	}
	if result.Data == nil || len(result.Data.Result) != 1 || result.Data.Result[0].Metric["job"] != "vm" {
		t.Fatalf("unexpected query result %+v", result.Data)
	}
	if !sawQueryRange {
		t.Fatal("expected registry client to hit fixture /api/v1/query_range")
	}

	labels, err := vm.Labels(ctx, "")
	if err != nil {
		t.Fatalf("Labels: %v", err)
	}
	if len(labels) != 2 || labels[0] != "__name__" || labels[1] != "job" {
		t.Fatalf("Labels=%v", labels)
	}
	if !sawLabels {
		t.Fatal("expected Labels to hit fixture /api/v1/labels")
	}

	names, err := vm.MetricNames(ctx, "")
	if err != nil {
		t.Fatalf("MetricNames: %v", err)
	}
	if len(names) != 2 || names[0] != "up" {
		t.Fatalf("MetricNames=%v", names)
	}
	if !sawMetricNames {
		t.Fatal("expected MetricNames to hit fixture /api/v1/label/__name__/values")
	}

	values, err := vm.LabelValues(ctx, "job", "")
	if err != nil {
		t.Fatalf("LabelValues: %v", err)
	}
	if len(values) != 2 || values[0] != "vm" {
		t.Fatalf("LabelValues=%v", values)
	}
	if !sawLabelValues {
		t.Fatal("expected LabelValues to hit fixture /api/v1/label/job/values")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawHealth {
		t.Fatal("expected TestConnection to hit fixture /health")
	}
}

func TestVictoriaMetrics_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{
		Type: models.DataSourceVictoriaMetrics,
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
