package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/janhoon/dash/backend/internal/auth"
)

func TestGoogleSSOConfigureRequiresAdmin(t *testing.T) {
	if testPool == nil {
		t.Skip("Database not available")
	}

	ctx := context.Background()

	// Create test org
	var orgID uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Test Org SSO', 'test-org-sso') RETURNING id`,
	).Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test org: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)

	// Create test user
	var userID uuid.UUID
	err = testPool.QueryRow(ctx,
		`INSERT INTO users (email, name) VALUES ('testssouser@example.com', 'Test SSO User') RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	// Add user as viewer (not admin)
	_, err = testPool.Exec(ctx,
		`INSERT INTO organization_memberships (user_id, organization_id, role) VALUES ($1, $2, 'viewer')`,
		userID, orgID,
	)
	if err != nil {
		t.Fatalf("Failed to add membership: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM organization_memberships WHERE user_id = $1`, userID)

	// Generate token for user
	token, err := testJWTManager.GenerateAccessToken(userID, "testssouser@example.com", "Test SSO User")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create handler
	handler := NewGoogleSSOHandler(testPool, testJWTManager)

	// Try to configure SSO as non-admin
	body := `{"client_id":"test-client-id","client_secret":"test-secret"}`
	req := httptest.NewRequest("POST", "/api/orgs/"+orgID.String()+"/sso/google", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", orgID.String())
	w := httptest.NewRecorder()

	wrappedHandler := auth.RequireAuth(testJWTManager, handler.ConfigureSSO)
	wrappedHandler(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-admin, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGoogleSSOConfigureAsAdmin(t *testing.T) {
	if testPool == nil {
		t.Skip("Database not available")
	}

	ctx := context.Background()

	// Create test org
	var orgID uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Test Org SSO Admin', 'test-org-sso-admin') RETURNING id`,
	).Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test org: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM sso_configs WHERE organization_id = $1`, orgID)
	defer testPool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)

	// Create test user
	var userID uuid.UUID
	err = testPool.QueryRow(ctx,
		`INSERT INTO users (email, name) VALUES ('testssoadmin@example.com', 'Test SSO Admin') RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	// Add user as admin
	_, err = testPool.Exec(ctx,
		`INSERT INTO organization_memberships (user_id, organization_id, role) VALUES ($1, $2, 'admin')`,
		userID, orgID,
	)
	if err != nil {
		t.Fatalf("Failed to add membership: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM organization_memberships WHERE user_id = $1`, userID)

	// Generate token for user
	token, err := testJWTManager.GenerateAccessToken(userID, "testssoadmin@example.com", "Test SSO Admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create handler
	handler := NewGoogleSSOHandler(testPool, testJWTManager)

	// Configure SSO as admin
	body := `{"client_id":"test-client-id","client_secret":"test-secret"}`
	req := httptest.NewRequest("POST", "/api/orgs/"+orgID.String()+"/sso/google", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", orgID.String())
	w := httptest.NewRecorder()

	wrappedHandler := auth.RequireAuth(testJWTManager, handler.ConfigureSSO)
	wrappedHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response GoogleSSOConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ClientID != "test-client-id" {
		t.Errorf("Expected client_id 'test-client-id', got '%s'", response.ClientID)
	}
	if !response.Enabled {
		t.Error("Expected SSO to be enabled by default")
	}
}

func TestGoogleSSOGetConfig(t *testing.T) {
	if testPool == nil {
		t.Skip("Database not available")
	}

	ctx := context.Background()

	// Create test org
	var orgID uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Test Org SSO Get', 'test-org-sso-get') RETURNING id`,
	).Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test org: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM sso_configs WHERE organization_id = $1`, orgID)
	defer testPool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)

	// Create SSO config
	_, err = testPool.Exec(ctx,
		`INSERT INTO sso_configs (organization_id, provider, client_id, client_secret, enabled)
		 VALUES ($1, 'google', 'get-client-id', 'get-secret', true)`,
		orgID,
	)
	if err != nil {
		t.Fatalf("Failed to create SSO config: %v", err)
	}

	// Create test user
	var userID uuid.UUID
	err = testPool.QueryRow(ctx,
		`INSERT INTO users (email, name) VALUES ('testssoadminget@example.com', 'Test SSO Admin Get') RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	// Add user as admin
	_, err = testPool.Exec(ctx,
		`INSERT INTO organization_memberships (user_id, organization_id, role) VALUES ($1, $2, 'admin')`,
		userID, orgID,
	)
	if err != nil {
		t.Fatalf("Failed to add membership: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM organization_memberships WHERE user_id = $1`, userID)

	// Generate token for user
	token, err := testJWTManager.GenerateAccessToken(userID, "testssoadminget@example.com", "Test SSO Admin Get")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create handler
	handler := NewGoogleSSOHandler(testPool, testJWTManager)

	// Get SSO config
	req := httptest.NewRequest("GET", "/api/orgs/"+orgID.String()+"/sso/google", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.SetPathValue("id", orgID.String())
	w := httptest.NewRecorder()

	wrappedHandler := auth.RequireAuth(testJWTManager, handler.GetSSOConfig)
	wrappedHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response GoogleSSOConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ClientID != "get-client-id" {
		t.Errorf("Expected client_id 'get-client-id', got '%s'", response.ClientID)
	}
}

func TestGoogleSSOLoginRequiresOrg(t *testing.T) {
	if testPool == nil {
		t.Skip("Database not available")
	}

	handler := NewGoogleSSOHandler(testPool, testJWTManager)

	// Try login without org parameter
	req := httptest.NewRequest("GET", "/api/auth/google/login", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing org, got %d", w.Code)
	}
}

func TestGoogleSSOLoginOrgNotConfigured(t *testing.T) {
	if testPool == nil {
		t.Skip("Database not available")
	}

	ctx := context.Background()

	// Create test org without SSO config
	var orgID uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Test Org No SSO', 'test-org-no-sso') RETURNING id`,
	).Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test org: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)

	handler := NewGoogleSSOHandler(testPool, testJWTManager)

	// Try login with org that doesn't have SSO configured
	req := httptest.NewRequest("GET", "/api/auth/google/login?org=test-org-no-sso", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for unconfigured SSO, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGoogleSSOLoginRedirectsToGoogle(t *testing.T) {
	if testPool == nil {
		t.Skip("Database not available")
	}

	ctx := context.Background()

	// Create test org with SSO config
	var orgID uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Test Org SSO Redirect', 'test-org-sso-redirect') RETURNING id`,
	).Scan(&orgID)
	if err != nil {
		t.Fatalf("Failed to create test org: %v", err)
	}
	defer testPool.Exec(ctx, `DELETE FROM sso_configs WHERE organization_id = $1`, orgID)
	defer testPool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)

	// Create SSO config
	_, err = testPool.Exec(ctx,
		`INSERT INTO sso_configs (organization_id, provider, client_id, client_secret, enabled)
		 VALUES ($1, 'google', 'redirect-client-id', 'redirect-secret', true)`,
		orgID,
	)
	if err != nil {
		t.Fatalf("Failed to create SSO config: %v", err)
	}

	handler := NewGoogleSSOHandler(testPool, testJWTManager)

	// Try login - should redirect to Google
	req := httptest.NewRequest("GET", "/api/auth/google/login?org=test-org-sso-redirect", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d: %s", w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Expected Location header for redirect")
	}

	// Check it's a Google URL
	if len(location) < 30 || location[:30] != "https://accounts.google.com/o" {
		t.Errorf("Expected redirect to Google, got: %s", location)
	}

	// Check state cookie was set
	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			stateCookie = c
			break
		}
	}
	if stateCookie == nil {
		t.Error("Expected oauth_state cookie to be set")
	}
}
