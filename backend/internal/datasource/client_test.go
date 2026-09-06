package datasource

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	acech "github.com/aceobservability/ace-datasource-clickhouse"
	acees "github.com/aceobservability/ace-datasource-elasticsearch"
	aceloki "github.com/aceobservability/ace-datasource-loki"
	aceprom "github.com/aceobservability/ace-datasource-prometheus"
	acetempo "github.com/aceobservability/ace-datasource-tempo"
	acevl "github.com/aceobservability/ace-datasource-victorialogs"
	acevm "github.com/aceobservability/ace-datasource-victoriametrics"
	acevt "github.com/aceobservability/ace-datasource-victoriatraces"

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
	if _, ok := client.(*aceprom.Client); !ok {
		t.Errorf("expected *prometheus.Client from ace-datasource-prometheus, got %T", client)
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
	if _, ok := client.(*acevm.Client); !ok {
		t.Errorf("expected *victoriametrics.Client from ace-datasource-victoriametrics, got %T", client)
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
	if _, ok := client.(*aceloki.Client); !ok {
		t.Errorf("expected *loki.Client from ace-datasource-loki, got %T", client)
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
	if _, ok := client.(*acevl.Client); !ok {
		t.Errorf("expected *victorialogs.Client from ace-datasource-victorialogs, got %T", client)
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
	if _, ok := client.(*acetempo.Client); !ok {
		t.Errorf("expected *tempo.Client from ace-datasource-tempo, got %T", client)
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
	if _, ok := client.(*acevt.Client); !ok {
		t.Errorf("expected *victoriatraces.Client from ace-datasource-victoriatraces, got %T", client)
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
	if _, ok := client.(*acech.Client); !ok {
		t.Errorf("expected *clickhouse.Client from ace-datasource-clickhouse, got %T", client)
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
	if _, ok := client.(*acees.Client); !ok {
		t.Errorf("expected *elasticsearch.Client from ace-datasource-elasticsearch, got %T", client)
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
	wantName := "probe-one"
	wantURL := "http://localhost:9090"
	wantAuthType := "bearer"
	wantAuthConfig := json.RawMessage(`{"token":"t"}`)
	wantTraceID := "trace_id"

	var got dscontract.Config
	dscontract.RegisterDatasource(typ, func(cfg dscontract.Config) (dscontract.Client, error) {
		got = cfg
		return &probeQueryClient{name: cfg.Name}, nil
	})
	t.Cleanup(func() { dscontract.UnregisterDatasource(typ) })

	client, err := NewClient(models.DataSource{
		Type:         models.DataSourceType(typ),
		Name:         wantName,
		URL:          wantURL,
		AuthType:     wantAuthType,
		AuthConfig:   wantAuthConfig,
		TraceIDField: wantTraceID,
	})
	if err != nil {
		t.Fatalf("dispatch registered type: %v", err)
	}
	gotClient, ok := client.(*probeQueryClient)
	if !ok {
		t.Fatalf("expected *probeQueryClient, got %T", client)
	}
	if gotClient.name != wantName {
		t.Errorf("factory did not receive Config.Name, name=%q", gotClient.name)
	}
	if got.Type != typ {
		t.Errorf("Type=%q, want %q", got.Type, typ)
	}
	if got.URL != wantURL {
		t.Errorf("URL=%q, want %q", got.URL, wantURL)
	}
	if got.AuthType != wantAuthType {
		t.Errorf("AuthType=%q, want %q", got.AuthType, wantAuthType)
	}
	if !bytes.Equal(got.AuthConfig, wantAuthConfig) {
		t.Errorf("AuthConfig=%s, want %s", got.AuthConfig, wantAuthConfig)
	}
	if got.TraceIDField != wantTraceID {
		t.Errorf("TraceIDField=%q, want %q", got.TraceIDField, wantTraceID)
	}
}

var (
	_ Client                  = (*aceprom.Client)(nil)
	_ MetricLabelsClient      = (*aceprom.Client)(nil)
	_ MetricLabelValuesClient = (*aceprom.Client)(nil)
	_ MetricNamesClient       = (*aceprom.Client)(nil)
	_ connectionTester        = (*aceprom.Client)(nil)
	_ httpClientProvider      = (*aceprom.Client)(nil)

	_ Client                  = (*acevm.Client)(nil)
	_ MetricLabelsClient      = (*acevm.Client)(nil)
	_ MetricLabelValuesClient = (*acevm.Client)(nil)
	_ MetricNamesClient       = (*acevm.Client)(nil)
	_ connectionTester        = (*acevm.Client)(nil)

	_ Client            = (*aceloki.Client)(nil)
	_ StreamClient      = (*aceloki.Client)(nil)
	_ LabelsClient      = (*aceloki.Client)(nil)
	_ LabelValuesClient = (*aceloki.Client)(nil)
	_ connectionTester  = (*aceloki.Client)(nil)

	_ Client             = (*acevl.Client)(nil)
	_ StreamClient       = (*acevl.Client)(nil)
	_ LabelsClient       = (*acevl.Client)(nil)
	_ LabelValuesClient  = (*acevl.Client)(nil)
	_ connectionTester   = (*acevl.Client)(nil)
	_ httpClientProvider = (*acevl.Client)(nil)

	_ Client            = (*acech.Client)(nil)
	_ SignalQueryClient = (*acech.Client)(nil)
	_ connectionTester  = (*acech.Client)(nil)

	_ Client            = (*CloudWatchClient)(nil)
	_ SignalQueryClient = (*CloudWatchClient)(nil)
	_ connectionTester  = (*CloudWatchClient)(nil)

	_ Client            = (*acees.Client)(nil)
	_ SignalQueryClient = (*acees.Client)(nil)
	_ connectionTester  = (*acees.Client)(nil)

	_ Client             = (*acetempo.Client)(nil)
	_ TracingClient      = (*acetempo.Client)(nil)
	_ connectionTester   = (*acetempo.Client)(nil)
	_ httpClientProvider = (*acetempo.Client)(nil)

	_ Client             = (*acevt.Client)(nil)
	_ TracingClient      = (*acevt.Client)(nil)
	_ connectionTester   = (*acevt.Client)(nil)
	_ httpClientProvider = (*acevt.Client)(nil)
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
