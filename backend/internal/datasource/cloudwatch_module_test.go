package datasource

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	acecw "github.com/aceobservability/ace-datasource-cloudwatch"

	"github.com/aceobservability/ace/backend/internal/models"
	"github.com/aceobservability/ace/backend/internal/ssrf"
)

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
	if _, isAuth := httpClient.Transport.(*dataSourceAuthRoundTripper); isAuth {
		t.Fatal("CloudWatch must not wrap DatasourceClient with dataSourceAuthRoundTripper (strips SigV4 on logs host)")
	}
	if httpClient.Transport == nil {
		t.Fatal("transport is nil")
	}
	if httpClient.Transport == http.DefaultTransport {
		t.Fatal("must not fall back to http.DefaultTransport")
	}
	policyType := reflect.TypeOf(ssrf.DatasourceClient(time.Second).Transport)
	if got := reflect.TypeOf(httpClient.Transport); got != policyType {
		t.Fatalf("Transport type %s, want DatasourceClient policy %s", got, policyType)
	}
}

func TestCloudWatchModule_PreservesSigV4OnCrossOrigin(t *testing.T) {
	t.Parallel()

	// ds.URL is the metrics amazonaws host. The request below is another
	// origin (logs.* in production; httptest here so CI does not dial AWS).
	// wrapDatasourceAuth would strip Authorization on that mismatch.
	const sigv4 = "AWS4-HMAC-SHA256 Credential=AKID/20260906/us-east-1/logs/aws4_request, SignedHeaders=host;x-amz-date, Signature=deadbeef"

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

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
		t.Fatalf("expected *cloudwatch.Client, got %T", client)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Authorization", sigv4)

	resp, err := cw.HTTPClient().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if gotAuth != sigv4 {
		t.Fatalf("Authorization after transport = %q, want SigV4 preserved on non-metrics origin", gotAuth)
	}
}
