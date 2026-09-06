// Package datasource is the stable import path for Ace datasource module contracts.
//
// External ace-datasource-* modules import
// github.com/aceobservability/ace/backend/pkg/datasource and call
// RegisterDatasource from init. Ace wires those factories from
// internal/datasource (SSRF injection). Do not import internal/datasource
// or internal/handlers from a module.
package datasource

import (
	"context"
	"time"
)

// Client is the interface every query datasource implements.
type Client interface {
	Query(ctx context.Context, query string, start, end time.Time, step time.Duration, limit int) (*QueryResult, error)
}

// SignalQueryClient is implemented by datasources that dispatch on signal
// (ClickHouse, CloudWatch, Elasticsearch).
type SignalQueryClient interface {
	QueryWithSignal(ctx context.Context, query, signal string, start, end time.Time, step time.Duration, limit int) (*QueryResult, error)
}

// StreamClient is implemented by log datasources that support live tail.
type StreamClient interface {
	Stream(ctx context.Context, query string, start time.Time, limit int, onLog LogStreamCallback) error
}

// LabelsClient is implemented by log datasources that expose label/field names.
type LabelsClient interface {
	Labels(ctx context.Context) ([]string, error)
}

// MetricLabelsClient is implemented by PromQL datasources that can scope labels
// to a metric selector.
type MetricLabelsClient interface {
	Labels(ctx context.Context, metric string) ([]string, error)
}

// LabelValuesClient is implemented by log datasources that expose label values.
type LabelValuesClient interface {
	LabelValues(ctx context.Context, labelName string) ([]string, error)
}

// MetricLabelValuesClient is implemented by PromQL datasources that can scope
// label values to a metric selector.
type MetricLabelValuesClient interface {
	LabelValues(ctx context.Context, label, metric string) ([]string, error)
}

// MetricNamesClient is implemented by PromQL datasources that expose metric names.
type MetricNamesClient interface {
	MetricNames(ctx context.Context, search string) ([]string, error)
}

// LogStreamCallback receives one live-tail log line.
type LogStreamCallback func(LogEntry) error

// QueryResult is the unified query result format.
type QueryResult struct {
	Status     string     `json:"status"`
	Data       *QueryData `json:"data,omitempty"`
	Error      string     `json:"error,omitempty"`
	ResultType string     `json:"resultType"` // "metrics" or "logs"
}

// QueryData contains the result.
type QueryData struct {
	ResultType string         `json:"resultType"`
	Result     []MetricResult `json:"result,omitempty"`
	Logs       []LogEntry     `json:"logs,omitempty"`
	Traces     []TraceSpan    `json:"traces,omitempty"`
}

// MetricResult is a single metric series (Prometheus/VictoriaMetrics).
type MetricResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

// LogEntry is a single log line (Loki/VictoriaLogs).
type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
	Level     string            `json:"level,omitempty"`
}

// TraceSpan is a normalized span in a query result.
type TraceSpan struct {
	SpanID            string            `json:"spanId"`
	ParentSpanID      string            `json:"parentSpanId,omitempty"`
	OperationName     string            `json:"operationName"`
	ServiceName       string            `json:"serviceName"`
	StartTimeUnixNano int64             `json:"startTimeUnixNano"`
	DurationNano      int64             `json:"durationNano"`
	Tags              map[string]string `json:"tags,omitempty"`
	Logs              []TraceLog        `json:"logs,omitempty"`
	Status            string            `json:"status,omitempty"`
}

// TraceLog is a normalized span log/event.
type TraceLog struct {
	TimestampUnixNano int64             `json:"timestampUnixNano"`
	Fields            map[string]string `json:"fields,omitempty"`
}
