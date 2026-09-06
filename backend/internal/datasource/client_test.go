package datasource

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/aceobservability/ace/backend/internal/models"
	dscontract "github.com/aceobservability/ace/backend/pkg/datasource"
)

type probeQueryClient struct {
	name string
}

func (c *probeQueryClient) Query(context.Context, string, time.Time, time.Time, time.Duration, int) (*QueryResult, error) {
	return &QueryResult{Status: c.name}, nil
}

func TestNewClient_Prometheus(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourcePrometheus,
		URL:  "http://localhost:9090",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*PrometheusClient); !ok {
		t.Errorf("expected PrometheusClient, got %T", client)
	}
}

func TestNewClient_VictoriaMetrics(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceVictoriaMetrics,
		URL:  "http://localhost:8428",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*VictoriaMetricsClient); !ok {
		t.Errorf("expected VictoriaMetricsClient, got %T", client)
	}
}

func TestNewClient_Loki(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceLoki,
		URL:  "http://localhost:3100",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*LokiClient); !ok {
		t.Errorf("expected LokiClient, got %T", client)
	}
}

func TestNewClient_VictoriaLogs(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceVictoriaLogs,
		URL:  "http://localhost:9428",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*VictoriaLogsClient); !ok {
		t.Errorf("expected VictoriaLogsClient, got %T", client)
	}
}

func TestNewClient_Tempo(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceTempo,
		URL:  "http://localhost:3200",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*TempoClient); !ok {
		t.Errorf("expected TempoClient, got %T", client)
	}
}

func TestNewClient_VictoriaTraces(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceVictoriaTraces,
		URL:  "http://localhost:10428",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*VictoriaTracesClient); !ok {
		t.Errorf("expected VictoriaTracesClient, got %T", client)
	}
}

func TestNewClient_ClickHouse(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceClickHouse,
		URL:  "http://localhost:8123",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*ClickHouseClient); !ok {
		t.Errorf("expected ClickHouseClient, got %T", client)
	}
}

func TestNewClient_CloudWatch(t *testing.T) {
	ds := models.DataSource{
		Type:       models.DataSourceCloudWatch,
		URL:        "https://monitoring.us-east-1.amazonaws.com",
		AuthConfig: []byte(`{"region":"us-east-1"}`),
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*CloudWatchClient); !ok {
		t.Errorf("expected CloudWatchClient, got %T", client)
	}
}

func TestNewClient_Elasticsearch(t *testing.T) {
	ds := models.DataSource{
		Type: models.DataSourceElasticsearch,
		URL:  "http://localhost:9200",
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*ElasticsearchClient); !ok {
		t.Errorf("expected ElasticsearchClient, got %T", client)
	}
}

func TestNewClient_InvalidType(t *testing.T) {
	ds := models.DataSource{
		Type: "invalid",
		URL:  "http://localhost:9090",
	}
	client, err := NewClient(ds)
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
	if !errors.Is(err, ErrUnknownType) {
		t.Fatalf("expected ErrUnknownType, got %v", err)
	}
	if client != nil {
		t.Fatalf("unknown type must not construct a client, got %T", client)
	}
}

func TestNewClient_VMAlertAndAlertManagerFailClosed(t *testing.T) {
	for _, typ := range []models.DataSourceType{models.DataSourceVMAlert, models.DataSourceAlertManager} {
		client, err := NewClient(models.DataSource{Type: typ, URL: "http://localhost:8880"})
		if !errors.Is(err, ErrUnknownType) {
			t.Errorf("%s: expected ErrUnknownType, got %v", typ, err)
		}
		if client != nil {
			t.Errorf("%s: unknown query type must not construct a client, got %T", typ, client)
		}
	}
}

func TestRegisterDatasource_DispatchUsesRegisteredFactory(t *testing.T) {
	const typ = "probe-internal-datasource"
	dscontract.RegisterDatasource(typ, func(cfg dscontract.Config) (dscontract.Client, error) {
		return &probeQueryClient{name: cfg.Name}, nil
	})
	t.Cleanup(func() { dscontract.UnregisterDatasource(typ) })

	client, err := NewClient(models.DataSource{
		Type: models.DataSourceType(typ),
		Name: "probe-one",
		URL:  "http://localhost:9090",
	})
	if err != nil {
		t.Fatalf("dispatch registered type: %v", err)
	}
	got, ok := client.(*probeQueryClient)
	if !ok {
		t.Fatalf("expected *probeQueryClient, got %T", client)
	}
	if got.name != "probe-one" {
		t.Errorf("factory did not receive Config, name=%q", got.name)
	}
}

var (
	_ Client                  = (*PrometheusClient)(nil)
	_ MetricLabelsClient      = (*PrometheusClient)(nil)
	_ MetricLabelValuesClient = (*PrometheusClient)(nil)
	_ MetricNamesClient       = (*PrometheusClient)(nil)
	_ connectionTester        = (*PrometheusClient)(nil)

	_ Client                  = (*VictoriaMetricsClient)(nil)
	_ MetricLabelsClient      = (*VictoriaMetricsClient)(nil)
	_ MetricLabelValuesClient = (*VictoriaMetricsClient)(nil)
	_ MetricNamesClient       = (*VictoriaMetricsClient)(nil)
	_ connectionTester        = (*VictoriaMetricsClient)(nil)

	_ Client            = (*LokiClient)(nil)
	_ StreamClient      = (*LokiClient)(nil)
	_ LabelsClient      = (*LokiClient)(nil)
	_ LabelValuesClient = (*LokiClient)(nil)
	_ connectionTester  = (*LokiClient)(nil)

	_ Client            = (*VictoriaLogsClient)(nil)
	_ StreamClient      = (*VictoriaLogsClient)(nil)
	_ LabelsClient      = (*VictoriaLogsClient)(nil)
	_ LabelValuesClient = (*VictoriaLogsClient)(nil)
	_ connectionTester  = (*VictoriaLogsClient)(nil)

	_ Client            = (*ClickHouseClient)(nil)
	_ SignalQueryClient = (*ClickHouseClient)(nil)
	_ connectionTester  = (*ClickHouseClient)(nil)

	_ Client            = (*CloudWatchClient)(nil)
	_ SignalQueryClient = (*CloudWatchClient)(nil)
	_ connectionTester  = (*CloudWatchClient)(nil)

	_ Client            = (*ElasticsearchClient)(nil)
	_ SignalQueryClient = (*ElasticsearchClient)(nil)
	_ connectionTester  = (*ElasticsearchClient)(nil)

	_ connectionTester = (*TempoClient)(nil)
	_ connectionTester = (*VictoriaTracesClient)(nil)
)

func TestDetectLogLevel(t *testing.T) {
	tests := []struct {
		labels map[string]string
		line   string
		want   string
	}{
		{map[string]string{"level": "ERROR"}, "some message", "error"},
		{map[string]string{"severity": "Warning"}, "some message", "warning"},
		{map[string]string{"severity": "Unspecified"}, "level=info msg=\"query\"", "info"},
		{map[string]string{"severity": "Unspecified"}, "> level=info ts=2026-02-08T14:30:26Z msg=\"query\"", "info"},
		{map[string]string{"severity_text": "ERROR2"}, "some message", "error"},
		{map[string]string{}, "Error: something failed", "error"},
		{map[string]string{}, "WARN: low disk space", "warning"},
		{map[string]string{}, "INFO starting service", "info"},
		{map[string]string{}, "DEBUG verbose output", "debug"},
		{map[string]string{}, "just a regular log line", ""},
	}

	for _, tt := range tests {
		got := detectLogLevel(tt.labels, tt.line)
		if got != tt.want {
			t.Errorf("detectLogLevel(%v, %q) = %q, want %q", tt.labels, tt.line, got, tt.want)
		}
	}
}

func TestToWebSocketURL(t *testing.T) {
	params := url.Values{}
	params.Set("query", `{job="api"}`)

	wsURL, err := toWebSocketURL("http://localhost:3100", "/loki/api/v1/tail", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsedURL, err := url.Parse(wsURL)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	if parsedURL.Scheme != "ws" {
		t.Fatalf("expected ws scheme, got %s", parsedURL.Scheme)
	}
	if parsedURL.Path != "/loki/api/v1/tail" {
		t.Fatalf("expected /loki/api/v1/tail path, got %s", parsedURL.Path)
	}
	if parsedURL.Query().Get("query") != `{job="api"}` {
		t.Fatalf("expected encoded query to round-trip, got %s", parsedURL.Query().Get("query"))
	}
}

func TestParseVictoriaLogsLine(t *testing.T) {
	entry, ok := parseVictoriaLogsLine(`{"_msg":"boom","_time":"2026-02-08T12:00:00Z","service":"api","level":"error"}`)
	if !ok {
		t.Fatal("expected line to parse")
	}

	if entry.Line != "boom" {
		t.Fatalf("expected line to be boom, got %q", entry.Line)
	}
	if entry.Timestamp != "2026-02-08T12:00:00Z" {
		t.Fatalf("expected timestamp to match, got %q", entry.Timestamp)
	}
	if entry.Labels["service"] != "api" {
		t.Fatalf("expected service label api, got %q", entry.Labels["service"])
	}
	if entry.Level != "error" {
		t.Fatalf("expected level error, got %q", entry.Level)
	}
}

func TestParseVictoriaLogsLineInvalid(t *testing.T) {
	if _, ok := parseVictoriaLogsLine(`not-json`); ok {
		t.Fatal("expected invalid line to fail parsing")
	}
}
