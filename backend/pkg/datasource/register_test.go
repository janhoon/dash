package datasource

import (
	"bytes"
	"context"
	"encoding/json"
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
	c, err := NewClient(Config{Type: "nope", URL: "http://localhost:9090"})
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
	want := Config{
		Type:         typ,
		Name:         "probe-one",
		URL:          "http://localhost:9090",
		AuthType:     "bearer",
		AuthConfig:   json.RawMessage(`{"token":"t"}`),
		TraceIDField: "trace_id",
	}
	var got Config
	RegisterDatasource(typ, func(cfg Config) (Client, error) {
		got = cfg
		return &stubClient{id: cfg.Name}, nil
	})
	t.Cleanup(func() { UnregisterDatasource(typ) })

	c, err := NewClient(want)
	if err != nil {
		t.Fatalf("dispatch registered type: %v", err)
	}
	gotClient, ok := c.(*stubClient)
	if !ok {
		t.Fatalf("expected *stubClient, got %T", c)
	}
	if gotClient.id != want.Name {
		t.Errorf("factory did not receive Config.Name, id=%q", gotClient.id)
	}
	if got.Type != want.Type {
		t.Errorf("Type=%q, want %q", got.Type, want.Type)
	}
	if got.URL != want.URL {
		t.Errorf("URL=%q, want %q", got.URL, want.URL)
	}
	if got.AuthType != want.AuthType {
		t.Errorf("AuthType=%q, want %q", got.AuthType, want.AuthType)
	}
	if !bytes.Equal(got.AuthConfig, want.AuthConfig) {
		t.Errorf("AuthConfig=%s, want %s", got.AuthConfig, want.AuthConfig)
	}
	if got.TraceIDField != want.TraceIDField {
		t.Errorf("TraceIDField=%q, want %q", got.TraceIDField, want.TraceIDField)
	}

	result, err := c.Query(context.Background(), "up", time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result.Status != want.Name {
		t.Fatalf("unexpected result status %q", result.Status)
	}
}

func TestRegisterDatasource_DuplicatePanics(t *testing.T) {
	const typ = "probe-dup-datasource"
	factory := func(Config) (Client, error) { return &stubClient{}, nil }
	RegisterDatasource(typ, factory)
	t.Cleanup(func() { UnregisterDatasource(typ) })

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate RegisterDatasource")
			return
		}
		got, ok := r.(string)
		if !ok {
			t.Fatalf("panic type %T, want string", r)
		}
		want := "RegisterDatasource: type already registered: " + typ
		if got != want {
			t.Fatalf("panic %q, want %q", got, want)
		}
	}()
	RegisterDatasource(typ, factory)
}
