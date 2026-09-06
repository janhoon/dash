package datasource

import (
	acetracing "github.com/aceobservability/ace-datasource-tempo/tracing"
	dscontract "github.com/aceobservability/ace/backend/pkg/datasource"
)

type (
	Client                  = dscontract.Client
	SignalQueryClient       = dscontract.SignalQueryClient
	StreamClient            = dscontract.StreamClient
	LabelsClient            = dscontract.LabelsClient
	MetricLabelsClient      = dscontract.MetricLabelsClient
	LabelValuesClient       = dscontract.LabelValuesClient
	MetricLabelValuesClient = dscontract.MetricLabelValuesClient
	MetricNamesClient       = dscontract.MetricNamesClient
	LogStreamCallback       = dscontract.LogStreamCallback
	QueryResult             = dscontract.QueryResult
	QueryData               = dscontract.QueryData
	MetricResult            = dscontract.MetricResult
	LogEntry                = dscontract.LogEntry
	TraceSpan               = dscontract.TraceSpan
	TraceLog                = dscontract.TraceLog
	TracingClient           = acetracing.TracingClient
	Trace                   = acetracing.Trace
	TraceSummary            = acetracing.TraceSummary
	TraceSearchRequest      = acetracing.TraceSearchRequest
	TraceServiceGraph       = acetracing.TraceServiceGraph
	TraceServiceNode        = acetracing.TraceServiceNode
	TraceServiceEdge        = acetracing.TraceServiceEdge
)

var ErrUnknownType = dscontract.ErrUnknownType
