package datasource

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	acech "github.com/aceobservability/ace-datasource-clickhouse"
	aceloki "github.com/aceobservability/ace-datasource-loki"
	aceprom "github.com/aceobservability/ace-datasource-prometheus"
	acevl "github.com/aceobservability/ace-datasource-victorialogs"
	acevm "github.com/aceobservability/ace-datasource-victoriametrics"

	"github.com/aceobservability/ace/backend/internal/models"
	"github.com/aceobservability/ace/backend/internal/ssrf"
)

func testDS(typ models.DataSourceType, rawURL string) models.DataSource {
	return models.DataSource{Type: typ, URL: rawURL}
}

const cloudMetadataURL = "http://169.254.169.254/"

func TestDatasourceClientsWireDialAndRedirectPolicy(t *testing.T) {
	t.Parallel()

	policyType := reflect.TypeOf(ssrf.DatasourceClient(time.Second).Transport)

	clients := constructedDatasourceHTTPClients(t)
	for name, client := range clients {
		t.Run(name, func(t *testing.T) {
			if client == nil {
				t.Fatal("http client is nil")
			}
			if client.CheckRedirect == nil {
				t.Fatal("expected DatasourceClient redirect policy (CheckRedirect)")
			}
			authRT, ok := client.Transport.(*dataSourceAuthRoundTripper)
			if !ok {
				t.Fatalf("Transport type %T, want *dataSourceAuthRoundTripper", client.Transport)
			}
			if authRT.base == nil {
				t.Fatal("auth round tripper inner transport is nil")
			}
			if authRT.base == http.DefaultTransport {
				t.Fatal("auth round tripper must not fall back to http.DefaultTransport")
			}
			if got := reflect.TypeOf(authRT.base); got != policyType {
				t.Fatalf("inner transport type %s, want DatasourceClient policy %s", got, policyType)
			}
		})
	}

	t.Run("victorialogs_stream_timeout", func(t *testing.T) {
		raw, err := NewClient(testDS(models.DataSourceVictoriaLogs, "http://127.0.0.1:9428"))
		if err != nil {
			t.Fatalf("NewClient victorialogs: %v", err)
		}
		c, ok := raw.(*acevl.Client)
		if !ok {
			t.Fatalf("expected *victorialogs.Client, got %T", raw)
		}
		if c.StreamHTTPClient() == nil {
			t.Fatal("stream client is nil")
		}
		if c.StreamHTTPClient().Timeout != 0 {
			t.Fatalf("streamClient.Timeout = %s, want 0 for long-lived tails", c.StreamHTTPClient().Timeout)
		}
	})
}

func TestDatasourceClientsAllowPrivateNetworks(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","data":{"result":[],"alerts":[],"groups":[]},"traces":[],"hits":{"hits":[]}}`))
	}))
	t.Cleanup(srv.Close)

	for name, query := range datasourceQueryFns() {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err := query(ctx, srv.URL)
			if isOutboundPolicyError(err) {
				t.Fatalf("query to private network should not be blocked by SSRF policy: %v", err)
			}
		})
	}
}

func TestDatasourceClientsBlockMetadata(t *testing.T) {
	t.Parallel()

	for name, query := range datasourceQueryFns() {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := query(ctx, strings.TrimRight(cloudMetadataURL, "/")); err == nil {
				t.Fatal("query to cloud metadata endpoint should fail")
			}
		})
	}
}

func TestDatasourceClientsBlockMetadataRedirect(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data", http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	for name, query := range datasourceQueryFns() {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := query(ctx, srv.URL); err == nil {
				t.Fatal("query with redirect to cloud metadata endpoint should fail")
			}
		})
	}
}

func TestLokiStreamBlocksMetadata(t *testing.T) {
	t.Parallel()

	client, err := NewClient(testDS(models.DataSourceLoki, strings.TrimRight(cloudMetadataURL, "/")))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	streamer, ok := client.(StreamClient)
	if !ok {
		t.Fatalf("expected StreamClient, got %T", client)
	}

	err = streamer.Stream(ctx, `{job="ace"}`, time.Time{}, 1, func(LogEntry) error { return nil })
	if err == nil {
		t.Fatal("Loki websocket stream to cloud metadata endpoint should fail")
	}
}

func TestVictoriaLogsStreamBlocksMetadata(t *testing.T) {
	t.Parallel()

	raw, err := NewClient(testDS(models.DataSourceVictoriaLogs, strings.TrimRight(cloudMetadataURL, "/")))
	if err != nil {
		t.Fatalf("NewClient victorialogs failed: %v", err)
	}
	client, ok := raw.(StreamClient)
	if !ok {
		t.Fatalf("expected StreamClient, got %T", raw)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Stream(ctx, `*`, time.Time{}, 1, func(LogEntry) error { return nil })
	if err == nil {
		t.Fatal("VictoriaLogs stream to cloud metadata endpoint should fail")
	}
}

func isOutboundPolicyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "cloud metadata") ||
		strings.Contains(msg, "private/internal") ||
		strings.Contains(msg, "redirect target rejected")
}

func constructedDatasourceHTTPClients(t *testing.T) map[string]*http.Client {
	t.Helper()

	prom, err := NewClient(testDS(models.DataSourcePrometheus, "http://127.0.0.1:9090"))
	if err != nil {
		t.Fatalf("NewClient prometheus: %v", err)
	}
	promClient, ok := prom.(*aceprom.Client)
	if !ok {
		t.Fatalf("expected *prometheus.Client, got %T", prom)
	}
	vm, err := NewClient(testDS(models.DataSourceVictoriaMetrics, "http://127.0.0.1:8428"))
	if err != nil {
		t.Fatalf("NewClient victoriametrics: %v", err)
	}
	vmClient, ok := vm.(*acevm.Client)
	if !ok {
		t.Fatalf("expected *victoriametrics.Client, got %T", vm)
	}
	loki, err := NewClient(testDS(models.DataSourceLoki, "http://127.0.0.1:3100"))
	if err != nil {
		t.Fatalf("NewClient loki: %v", err)
	}
	lokiClient, ok := loki.(*aceloki.Client)
	if !ok {
		t.Fatalf("expected *loki.Client, got %T", loki)
	}
	vlogsRaw, err := NewClient(testDS(models.DataSourceVictoriaLogs, "http://127.0.0.1:9428"))
	if err != nil {
		t.Fatalf("NewClient victorialogs: %v", err)
	}
	vlogs, ok := vlogsRaw.(*acevl.Client)
	if !ok {
		t.Fatalf("expected *victorialogs.Client, got %T", vlogsRaw)
	}
	tempo, err := NewTempoClient(models.DataSource{URL: "http://127.0.0.1:3200", Type: models.DataSourceTempo})
	if err != nil {
		t.Fatalf("NewTempoClient: %v", err)
	}
	vtraces, err := NewVictoriaTracesClient(models.DataSource{URL: "http://127.0.0.1:10428", Type: models.DataSourceVictoriaTraces})
	if err != nil {
		t.Fatalf("NewVictoriaTracesClient: %v", err)
	}
	ch, err := NewClient(models.DataSource{URL: "http://127.0.0.1:8123", Type: models.DataSourceClickHouse})
	if err != nil {
		t.Fatalf("NewClient clickhouse: %v", err)
	}
	chClient, ok := ch.(*acech.Client)
	if !ok {
		t.Fatalf("expected *clickhouse.Client, got %T", ch)
	}
	es, err := NewElasticsearchClient(models.DataSource{URL: "http://127.0.0.1:9200", Type: models.DataSourceElasticsearch})
	if err != nil {
		t.Fatalf("NewElasticsearchClient: %v", err)
	}
	am, err := NewAlertManagerClient(models.DataSource{URL: "http://127.0.0.1:9093", Type: models.DataSourceAlertManager})
	if err != nil {
		t.Fatalf("NewAlertManagerClient: %v", err)
	}
	vmalert, err := NewVMAlertClient(models.DataSource{URL: "http://127.0.0.1:8880", Type: models.DataSourceVMAlert})
	if err != nil {
		t.Fatalf("NewVMAlertClient: %v", err)
	}

	return map[string]*http.Client{
		"prometheus":          promClient.HTTPClient(),
		"victoriametrics":     vmClient.HTTPClient(),
		"loki":                lokiClient.HTTPClient(),
		"victorialogs":        vlogs.HTTPClient(),
		"victorialogs_stream": vlogs.StreamHTTPClient(),
		"tempo":               tempo.httpClient,
		"victoriatraces":      vtraces.httpClient,
		"clickhouse":          chClient.HTTPClient(),
		"elasticsearch":       es.httpClient,
		"alertmanager":        am.HTTPClient(),
		"vmalert":             vmalert.HTTPClient(),
	}
}

func datasourceQueryFns() map[string]func(context.Context, string) error {
	return map[string]func(context.Context, string) error{
		"prometheus": func(ctx context.Context, baseURL string) error {
			client, err := NewClient(testDS(models.DataSourcePrometheus, baseURL))
			if err != nil {
				return err
			}
			result, err := client.Query(ctx, "up", time.Now().Add(-time.Hour), time.Now(), time.Minute, 10)
			if err != nil {
				return err
			}
			if result != nil && result.Status == "error" && result.Error != "" {
				return fmt.Errorf("%s", result.Error)
			}
			return nil
		},
		"victoriametrics": func(ctx context.Context, baseURL string) error {
			client, err := NewClient(testDS(models.DataSourceVictoriaMetrics, baseURL))
			if err != nil {
				return err
			}
			_, err = client.Query(ctx, "up", time.Now().Add(-time.Hour), time.Now(), time.Minute, 10)
			return err
		},
		"loki": func(ctx context.Context, baseURL string) error {
			client, err := NewClient(testDS(models.DataSourceLoki, baseURL))
			if err != nil {
				return err
			}
			_, err = client.Query(ctx, `{job="ace"}`, time.Now().Add(-time.Hour), time.Now(), time.Minute, 10)
			return err
		},
		"victorialogs": func(ctx context.Context, baseURL string) error {
			client, err := NewClient(testDS(models.DataSourceVictoriaLogs, baseURL))
			if err != nil {
				return err
			}
			_, err = client.Query(ctx, `*`, time.Now().Add(-time.Hour), time.Now(), time.Minute, 10)
			return err
		},
		"tempo": func(ctx context.Context, baseURL string) error {
			client, err := NewTempoClient(models.DataSource{URL: baseURL, Type: models.DataSourceTempo})
			if err != nil {
				return err
			}
			_, err = client.GetTrace(ctx, "abc")
			return err
		},
		"victoriatraces": func(ctx context.Context, baseURL string) error {
			client, err := NewVictoriaTracesClient(models.DataSource{URL: baseURL, Type: models.DataSourceVictoriaTraces})
			if err != nil {
				return err
			}
			_, err = client.GetTrace(ctx, "abc")
			return err
		},
		"clickhouse": func(ctx context.Context, baseURL string) error {
			client, err := NewClient(testDS(models.DataSourceClickHouse, baseURL))
			if err != nil {
				return err
			}
			_, err = client.Query(ctx, "SELECT 1", time.Now().Add(-time.Hour), time.Now(), time.Minute, 10)
			return err
		},
		"elasticsearch": func(ctx context.Context, baseURL string) error {
			client, err := NewElasticsearchClient(models.DataSource{URL: baseURL, Type: models.DataSourceElasticsearch})
			if err != nil {
				return err
			}
			_, err = client.Query(ctx, "*", time.Now().Add(-time.Hour), time.Now(), time.Minute, 10)
			return err
		},
		"alertmanager": func(ctx context.Context, baseURL string) error {
			client, err := NewAlertManagerClient(models.DataSource{URL: baseURL, Type: models.DataSourceAlertManager})
			if err != nil {
				return err
			}
			_, err = client.GetStatus(ctx)
			return err
		},
		"vmalert": func(ctx context.Context, baseURL string) error {
			client, err := NewVMAlertClient(models.DataSource{URL: baseURL, Type: models.DataSourceVMAlert})
			if err != nil {
				return err
			}
			return client.Health(ctx)
		},
	}
}
