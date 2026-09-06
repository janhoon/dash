package datasource

import dscontract "github.com/aceobservability/ace/backend/pkg/datasource"

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
)

var ErrUnknownType = dscontract.ErrUnknownType
