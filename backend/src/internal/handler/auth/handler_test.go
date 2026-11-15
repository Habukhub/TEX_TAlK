package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tex/src/internal/configs"
	"tex/src/internal/domain"
	authhandler "tex/src/internal/handler/auth"
	"tex/src/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// --- Mock AuthUseCase ---
type mockAuthUseCase struct {
	mock.Mock
}

func (m *mockAuthUseCase) HandleGoogleCallback(ctx context.Context, googleID, email, name string) (*domain.User, string, error) {
	args := m.Called(ctx, googleID, email, name)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

var _ usecase.AuthUseCase = (*mockAuthUseCase)(nil)

// setupAuthApp สร้าง Fiber app ปกติ (Google Endpoint จริง)
func setupAuthApp(authUC usecase.AuthUseCase, conf *configs.Config) *fiber.App {
	app := fiber.New()
	h := authhandler.NewHandler(authUC, conf)
	app.Get("/auth/google", h.GoogleLogin)
	app.Get("/auth/google/callback", h.GoogleCallback)
	app.Post("/auth/logout", h.Logout)
	return app
}

// setupAuthAppWithOAuthEndpoint สร้าง Fiber app โดยใช้ Custom Endpoint (local test server)
func setupAuthAppWithOAuthEndpoint(authUC usecase.AuthUseCase, conf *configs.Config, endpoint oauth2.Endpoint) *fiber.App {
	app := fiber.New()
	h := authhandler.NewHandlerWithEndpoint(authUC, conf, endpoint)
	app.Get("/auth/google", h.GoogleLogin)
	app.Get("/auth/google/callback", h.GoogleCallback)
	app.Post("/auth/logout", h.Logout)
	return app
}

func testConfig() *configs.Config {
	return &configs.Config{
		Auth: configs.AuthConfig{
			GoogleClientID:     "test-client-id",
			GoogleClientSecret: "test-client-secret",
			GoogleRedirectURL:  "http://localhost:8080/auth/google/callback",
			JWTSecret:          "test-jwt-secret",
		},
	}
}

// ===========================================
// GoogleLogin Tests
// ===========================================

func TestGoogleLoginHandler_RedirectsToGoogle(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	app := setupAuthApp(mockUC, testConfig())

	req := httptest.NewRequest("GET", "/auth/google", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Location"))
}

// ===========================================
// GoogleCallback Tests
// ===========================================

func TestGoogleCallbackHandler_MissingState(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	app := setupAuthApp(mockUC, testConfig())

	// No state param and no cookie → redirects with error
	req := httptest.NewRequest("GET", "/auth/google/callback?code=fake-code", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "invalid_state")
	mockUC.AssertNotCalled(t, "HandleGoogleCallback")
}

func TestGoogleCallbackHandler_StateMismatch(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	app := setupAuthApp(mockUC, testConfig())

	req := httptest.NewRequest("GET", "/auth/google/callback?code=fake-code&state=aaa", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "bbb"})
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "invalid_state")
	mockUC.AssertNotCalled(t, "HandleGoogleCallback")
}

func TestGoogleCallbackHandler_MissingCode(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	app := setupAuthApp(mockUC, testConfig())

	// State matches but no code
	req := httptest.NewRequest("GET", "/auth/google/callback?state=mystate", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "mystate"})
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockUC.AssertNotCalled(t, "HandleGoogleCallback")
}

func TestGoogleCallbackHandler_TokenExchangeFailed(t *testing.T) {
	fakeTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad_request", http.StatusBadRequest)
	}))
	defer fakeTokenServer.Close()

	mockUC := new(mockAuthUseCase)
	endpoint := oauth2.Endpoint{
		TokenURL: fakeTokenServer.URL + "/token",
		AuthURL:  fakeTokenServer.URL + "/auth",
	}
	app := setupAuthAppWithOAuthEndpoint(mockUC, testConfig(), endpoint)

	req := httptest.NewRequest("GET", "/auth/google/callback?code=fake-code&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "teststate"})
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGoogleCallbackHandler_UserInfoFailed(t *testing.T) {
	fakeUserInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer fakeUserInfoServer.Close()

	fakeTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "fake-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "fake-refresh-token",
		})
	}))
	defer fakeTokenServer.Close()

	mockUC := new(mockAuthUseCase)
	endpoint := oauth2.Endpoint{
		TokenURL: fakeTokenServer.URL + "/token",
		AuthURL:  fakeTokenServer.URL + "/auth",
	}
	conf := testConfig()
	conf.Auth.UserInfoURL = fakeUserInfoServer.URL + "/userinfo"
	app := setupAuthAppWithOAuthEndpoint(mockUC, conf, endpoint)

	req := httptest.NewRequest("GET", "/auth/google/callback?code=fake-code&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "teststate"})
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGoogleCallbackHandler_FullSuccess(t *testing.T) {
	userInfo := map[string]interface{}{
		"id":    "google-user-123",
		"email": "user@example.com",
		"name":  "Test User",
	}

	// จำลอง UserInfo Server
	fakeUserInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer fakeUserInfoServer.Close()

	// จำลอง Token Server
	fakeTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "fake-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer fakeTokenServer.Close()

	mockUC := new(mockAuthUseCase)
	expectedUser := &domain.User{ID: uuid.New(), Email: "user@example.com"}
	mockUC.On("HandleGoogleCallback", mock.Anything, "google-user-123", "user@example.com", "Test User").
		Return(expectedUser, "jwt-token", nil)

	endpoint := oauth2.Endpoint{
		TokenURL: fakeTokenServer.URL + "/token",
		AuthURL:  fakeTokenServer.URL + "/auth",
	}
	conf := testConfig()
	conf.Auth.UserInfoURL = fakeUserInfoServer.URL + "/userinfo"

	app := setupAuthAppWithOAuthEndpoint(mockUC, conf, endpoint)

	req := httptest.NewRequest("GET", "/auth/google/callback?code=fake-code&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "teststate"})
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	mockUC.AssertExpectations(t)
}

// ===========================================
// Logout Tests
// ===========================================

func TestLogoutHandler_ReturnsOK(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	app := setupAuthApp(mockUC, testConfig())

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLogoutHandler_ReturnsMessage(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	app := setupAuthApp(mockUC, testConfig())

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)

	buf := make([]byte, resp.ContentLength)
	resp.Body.Read(buf)
	assert.Contains(t, string(buf), "logout successful")
}

// ===========================================
// Mock UseCase Tests
// ===========================================

func TestMockAuthUseCase_HandleGoogleCallback(t *testing.T) {
	mockUC := new(mockAuthUseCase)

	expectedUser := &domain.User{
		ID:       uuid.New(),
		GoogleID: "g-123",
		Email:    "user@example.com",
		Name:     "Test User",
	}

	mockUC.On("HandleGoogleCallback", mock.Anything, "g-123", "user@example.com", "Test User").
		Return(expectedUser, "jwt-token-string", nil)

	user, token, err := mockUC.HandleGoogleCallback(
		context.Background(), "g-123", "user@example.com", "Test User",
	)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "jwt-token-string", token)
	assert.Equal(t, "user@example.com", user.Email)
	mockUC.AssertExpectations(t)
}
