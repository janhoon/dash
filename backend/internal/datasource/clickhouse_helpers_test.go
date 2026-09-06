package datasource

import (
	"testing"
)

func TestNormaliseToMetrics(t *testing.T) {
	rows := []map[string]interface{}{
		{"timestamp": 1700000000, "value": 2.5, "host": "a", "metric_name": "cpu_usage"},
		{"timestamp": 1700000060, "value": 2.8, "host": "a", "metric_name": "cpu_usage"},
		{"timestamp": 1700000000, "value": 3.1, "host": "b", "metric_name": "cpu_usage"},
	}

	metrics := NormaliseToMetrics(rows)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metric series, got %d", len(metrics))
	}

	seriesByHost := map[string]MetricResult{}
	for _, series := range metrics {
		seriesByHost[series.Metric["host"]] = series
	}

	seriesA, ok := seriesByHost["a"]
	if !ok {
		t.Fatalf("expected host=a series to exist")
	}
	if len(seriesA.Values) != 2 {
		t.Fatalf("expected host=a to have 2 values, got %d", len(seriesA.Values))
	}
	if seriesA.Metric["__name__"] != "cpu_usage" {
		t.Fatalf("expected metric name cpu_usage, got %q", seriesA.Metric["__name__"])
	}

	firstTimestamp, ok := parseClickHouseFloat(seriesA.Values[0][0])
	if !ok || firstTimestamp != 1700000000 {
		t.Fatalf("expected first timestamp 1700000000, got %v", seriesA.Values[0][0])
	}
}
