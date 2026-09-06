package datasource

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// ErrUnknownType is returned when a datasource type has no registered factory.
// Callers must fail closed.
var ErrUnknownType = errors.New("unknown datasource type")

// Config is the in-process config passed to a datasource factory.
// It is the module-facing subset of a stored datasource row.
type Config struct {
	Name         string
	Type         string
	URL          string
	AuthType     string
	AuthConfig   json.RawMessage
	TraceIDField string
}

// Factory constructs a Client from Config.
type Factory func(Config) (Client, error)

var registry = struct {
	mu sync.RWMutex
	m  map[string]Factory
}{
	m: map[string]Factory{},
}

// RegisterDatasource records a factory for datasource type.
func RegisterDatasource(typ string, factory Factory) {
	if typ == "" {
		panic("RegisterDatasource: empty type")
	}
	if factory == nil {
		panic("RegisterDatasource: nil factory")
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.m[typ] = factory
}

func lookup(typ string) (Factory, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	f, ok := registry.m[typ]
	return f, ok
}

// UnregisterDatasource removes a factory. Tests use this to isolate probe types.
func UnregisterDatasource(typ string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.m, typ)
}

// NewClient constructs a client from the registry. Unknown types fail closed.
func NewClient(typ string, cfg Config) (Client, error) {
	factory, ok := lookup(typ)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownType, typ)
	}
	return factory(cfg)
}
