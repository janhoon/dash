package datasource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aceobservability/ace/backend/internal/models"
)

func TestVMAlert_TestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawHealth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			sawHealth = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{Type: models.DataSourceVMAlert, URL: srv.URL}
	if _, err := NewVMAlertClient(ds); err != nil {
		t.Fatalf("NewVMAlertClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawHealth {
		t.Fatal("expected TestConnection to hit fixture /health")
	}
}

func TestAlertManager_TestConnectionAgainstFixtureHTTP(t *testing.T) {
	t.Parallel()

	var sawStatus bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/status" {
			sawStatus = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"cluster":{"status":"ready"},"versionInfo":{"version":"0.27.0"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	ds := models.DataSource{Type: models.DataSourceAlertManager, URL: srv.URL}
	if _, err := NewAlertManagerClient(ds); err != nil {
		t.Fatalf("NewAlertManagerClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := TestConnection(ctx, ds); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !sawStatus {
		t.Fatal("expected TestConnection to hit fixture /api/v2/status")
	}
}

func TestVMAlert_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{Type: models.DataSourceVMAlert, URL: "http://169.254.169.254/"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := TestConnection(ctx, ds)
	if err == nil {
		t.Fatal("expected TestConnection to reject cloud metadata URL")
	}
	if !strings.Contains(err.Error(), "datasource url rejected") {
		t.Fatalf("error %q, want datasource url rejected", err)
	}
}

func TestAlertManager_TestConnectionRejectsMetadataURL(t *testing.T) {
	t.Parallel()

	ds := models.DataSource{Type: models.DataSourceAlertManager, URL: "http://169.254.169.254/"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := TestConnection(ctx, ds)
	if err == nil {
		t.Fatal("expected TestConnection to reject cloud metadata URL")
	}
	if !strings.Contains(err.Error(), "datasource url rejected") {
		t.Fatalf("error %q, want datasource url rejected", err)
	}
}
