package datasource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	aceprom "github.com/aceobservability/ace-datasource-prometheus"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestPrometheusModule_QueryAndTestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawQueryRange, sawHealthy bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/api/v1/query_range"):
			sawQueryRange = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","job":"prometheus"},"values":[[1600000000,"1"]]}]}}`))
		case r.URL.Path == "/-/healthy":
			sawHealthy = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{
		Type: models.DataSourcePrometheus,
		URL:  srv.URL,
	}

	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, ok := client.(*aceprom.Client); !ok {
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
	if result.Data == nil || len(result.Data.Result) != 1 || result.Data.Result[0].Metric["job"] != "prometheus" {
		t.Fatalf("unexpected query result %+v", result.Data)
	}
	if !sawQueryRange {
		t.Fatal("expected registry client to hit fixture /api/v1/query_range")
	}

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawHealthy {
		t.Fatal("expected TestConnection to hit fixture /-/healthy")
	}
}
