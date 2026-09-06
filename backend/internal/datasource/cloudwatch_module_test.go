package datasource

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	acecw "github.com/aceobservability/ace-datasource-cloudwatch"

	"github.com/aceobservability/ace/backend/internal/models"
	"github.com/aceobservability/ace/backend/internal/ssrf"
)

// CloudWatchClient aliases the extracted module client so existing compile-time
// assertions in this package keep compiling without editing shared type files.
type CloudWatchClient = acecw.Client

func TestCloudWatchModule_InjectsDatasourceHTTPClient(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{
		Type:       models.DataSourceCloudWatch,
		URL:        "https://monitoring.us-east-1.amazonaws.com",
		AuthConfig: json.RawMessage(`{"region":"us-east-1","access_key_id":"AKID","secret_access_key":"SECRET"}`),
	}
	client, err := NewClient(ds)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cw, ok := client.(*acecw.Client)
	if !ok {
		t.Fatalf("expected *cloudwatch.Client from ace-datasource-cloudwatch, got %T", client)
	}

	httpClient := cw.HTTPClient()
	if httpClient == nil {
		t.Fatal("http client is nil")
	}
	if httpClient.CheckRedirect == nil {
		t.Fatal("expected DatasourceClient redirect policy (CheckRedirect)")
	}
	authRT, ok := httpClient.Transport.(*dataSourceAuthRoundTripper)
	if !ok {
		t.Fatalf("Transport type %T, want *dataSourceAuthRoundTripper", httpClient.Transport)
	}
	if authRT.base == nil {
		t.Fatal("auth round tripper inner transport is nil")
	}
	if authRT.base == http.DefaultTransport {
		t.Fatal("auth round tripper must not fall back to http.DefaultTransport")
	}
	policyType := reflect.TypeOf(ssrf.DatasourceClient(time.Second).Transport)
	if got := reflect.TypeOf(authRT.base); got != policyType {
		t.Fatalf("inner transport type %s, want DatasourceClient policy %s", got, policyType)
	}
}
