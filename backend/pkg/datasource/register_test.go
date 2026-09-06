package datasource

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubClient struct {
	id string
}

func (s *stubClient) Query(context.Context, string, time.Time, time.Time, time.Duration, int) (*QueryResult, error) {
	return &QueryResult{Status: s.id, ResultType: "metrics"}, nil
}

func TestNewClient_UnknownTypeFailsClosed(t *testing.T) {
	c, err := NewClient("nope", Config{URL: "http://localhost:9090"})
	if err == nil {
		t.Fatal("expected error for unknown datasource type")
	}
	if !errors.Is(err, ErrUnknownType) {
		t.Fatalf("expected ErrUnknownType, got %v", err)
	}
	if c != nil {
		t.Fatalf("unknown type must not construct a client, got %T", c)
	}
}

func TestRegisterDatasource_DispatchUsesRegisteredFactory(t *testing.T) {
	const typ = "probe-pkg-datasource"
	RegisterDatasource(typ, func(cfg Config) (Client, error) {
		return &stubClient{id: cfg.Name}, nil
	})
	t.Cleanup(func() { UnregisterDatasource(typ) })

	c, err := NewClient(typ, Config{Name: "probe-one", URL: "http://localhost:9090"})
	if err != nil {
		t.Fatalf("dispatch registered type: %v", err)
	}
	got, ok := c.(*stubClient)
	if !ok {
		t.Fatalf("expected *stubClient, got %T", c)
	}
	if got.id != "probe-one" {
		t.Errorf("factory did not receive Config, id=%q", got.id)
	}

	result, err := c.Query(context.Background(), "up", time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result.Status != "probe-one" {
		t.Fatalf("unexpected result status %q", result.Status)
	}
}
