package datasource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	aceprom "github.com/aceobservability/ace-datasource-prometheus"
	acevm "github.com/aceobservability/ace-datasource-victoriametrics"

	"github.com/aceobservability/ace/backend/internal/models"
	"github.com/aceobservability/ace/backend/internal/ssrf"
	dscontract "github.com/aceobservability/ace/backend/pkg/datasource"
)

// QueryRequest represents a query request body
type QueryRequest struct {
	Query  string `json:"query"`
	Signal string `json:"signal,omitempty"`
	Start  int64  `json:"start"` // Unix timestamp in seconds
	End    int64  `json:"end"`   // Unix timestamp in seconds
	Step   int64  `json:"step"`  // Step interval in seconds
	Limit  int    `json:"limit"` // Max results for log queries
}

// StreamRequest represents a live stream request body
type StreamRequest struct {
	Query string `json:"query"`
	Start int64  `json:"start,omitempty"` // Unix timestamp in seconds for resume cursor
	Limit int    `json:"limit,omitempty"` // Max entries per tail batch
}

type connectionTester interface {
	TestConnection(ctx context.Context) error
}

// NewClient creates a datasource client based on the datasource type.
func NewClient(ds models.DataSource) (Client, error) {
	return dscontract.NewClient(configFromDataSource(ds))
}

func TestConnection(ctx context.Context, ds models.DataSource) error {
	switch ds.Type {
	case models.DataSourceVMAlert:
		client, err := NewVMAlertClient(ds)
		if err != nil {
			return err
		}
		return runHTTPConnectionCheck(ctx, ds, client.client, []string{"/health", "/api/v1/alerts", "/"})
	case models.DataSourceAlertManager:
		client, err := NewAlertManagerClient(ds)
		if err != nil {
			return err
		}
		return runHTTPConnectionCheck(ctx, ds, client.client, []string{"/api/v2/status", "/api/v2/alerts", "/"})
	case models.DataSourcePrometheus:
		client, err := NewClient(ds)
		if err != nil {
			return err
		}
		prom, ok := client.(*aceprom.Client)
		if !ok {
			return fmt.Errorf("prometheus client type %T", client)
		}
		return runHTTPConnectionCheck(ctx, ds, prom.HTTPClient(), []string{"/-/healthy", "/api/v1/query?query=1", "/"})
	case models.DataSourceVictoriaMetrics:
		client, err := NewClient(ds)
		if err != nil {
			return err
		}
		vm, ok := client.(*acevm.Client)
		if !ok {
			return fmt.Errorf("victoriametrics client type %T", client)
		}
		return runHTTPConnectionCheck(ctx, ds, vm.HTTPClient(), []string{"/health", "/api/v1/query?query=1", "/"})
	default:
		client, err := NewClient(ds)
		if err != nil {
			return err
		}
		tester, ok := client.(connectionTester)
		if !ok {
			return fmt.Errorf("unsupported datasource type: %s", ds.Type)
		}
		return tester.TestConnection(ctx)
	}
}

func runHTTPConnectionCheck(ctx context.Context, ds models.DataSource, httpClient *http.Client, endpoints []string) error {
	// Datasource URLs may legitimately target private/internal networks
	// (local Victoria stack, in-cluster Prometheus, etc.). Match create/update
	// policy: only reject non-http(s) and cloud metadata.
	//
	// Positive IsLocalURL guard: CodeQL RedirectCheckBarrier sanitizes ds.URL
	// on the true branch (function name matches isLocalUrl pattern). Keep the
	// request sinks inside this branch so the barrier applies.
	if !ssrf.IsLocalURL(ds.URL) {
		if _, err := ssrf.ValidateDatasourceURL(ds.URL); err != nil {
			return fmt.Errorf("datasource url rejected: %w", err)
		}
		return fmt.Errorf("datasource url rejected")
	}

	if httpClient == nil {
		httpClient = newDatasourceHTTPClient(ds, 10*time.Second)
	}

	baseURL := ds.URL

	var lastErr error
	for _, endpoint := range endpoints {
		targetURL, err := resolveHealthEndpoint(baseURL, endpoint)
		if err != nil {
			lastErr = err
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("endpoint %s not found", endpoint)
			continue
		}

		message := strings.TrimSpace(string(body))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}

		lastErr = fmt.Errorf("endpoint %s returned status %d: %s", endpoint, resp.StatusCode, message)
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("connection test failed")
}

func resolveHealthEndpoint(baseURL, endpoint string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid datasource url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported datasource url scheme: %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("datasource url has no host")
	}

	resolved, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid health endpoint %q: %w", endpoint, err)
	}

	return parsed.ResolveReference(resolved).String(), nil
}
