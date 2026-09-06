package datasource

import (
	"fmt"

	acetracing "github.com/aceobservability/ace-datasource-tempo/tracing"

	"github.com/aceobservability/ace/backend/internal/models"
)

// NewTracingClient returns a tracing client from the datasource registry.
func NewTracingClient(ds models.DataSource) (TracingClient, error) {
	client, err := NewClient(ds)
	if err != nil {
		return nil, err
	}
	tc, ok := client.(TracingClient)
	if !ok {
		return nil, fmt.Errorf("unsupported tracing datasource type: %s", ds.Type)
	}
	return tc, nil
}

// BuildTraceServiceGraph aggregates spans into service-level dependencies.
func BuildTraceServiceGraph(trace *Trace) *TraceServiceGraph {
	return acetracing.BuildTraceServiceGraph(trace)
}
