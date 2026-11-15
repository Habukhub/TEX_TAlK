package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"tex/src/internal/configs"
	"tex/src/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Handler struct {
	authUC      usecase.AuthUseCase
	oauthConfig *oauth2.Config
	userInfoURL string
	frontendURL string
}

const defaultUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

func NewHandler(authUC usecase.AuthUseCase, conf *configs.Config) *Handler {
	oauthConfig := &oauth2.Config{
		ClientID:     conf.Auth.GoogleClientID,
		ClientSecret: conf.Auth.GoogleClientSecret,
		RedirectURL:  conf.Auth.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	userInfoURL := conf.Auth.UserInfoURL
	if userInfoURL == "" {
		userInfoURL = defaultUserInfoURL
	}
	return &Handler{
		authUC:      authUC,
		oauthConfig: oauthConfig,
		userInfoURL: userInfoURL,
		frontendURL: strings.TrimRight(conf.App.FrontendBaseURL, "/"),
	}
}

// NewHandlerWithEndpoint สำหรับ Testing เท่านั้น — ให้ inject custom OAuth Endpoint ได้
func NewHandlerWithEndpoint(authUC usecase.AuthUseCase, conf *configs.Config, endpoint oauth2.Endpoint) *Handler {
	oauthConfig := &oauth2.Config{
		ClientID:     conf.Auth.GoogleClientID,
		ClientSecret: conf.Auth.GoogleClientSecret,
		RedirectURL:  conf.Auth.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: endpoint,
	}
	userInfoURL := conf.Auth.UserInfoURL
	if userInfoURL == "" {
		userInfoURL = defaultUserInfoURL
	}
	return &Handler{
		authUC:      authUC,
		oauthConfig: oauthConfig,
		userInfoURL: userInfoURL,
		frontendURL: strings.TrimRight(conf.App.FrontendBaseURL, "/"),
	}
}

// GoogleLogin godoc
// @Summary Google Login
// @Description Redirect to Google OAuth2 login page
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 302 {string} string "Redirect"
// @Router /api/v1/auth/google [get]
func (h *Handler) GoogleLogin(c fiber.Ctx) error {
	state := generateState()
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		MaxAge:   300,
		Path:     "/",
	})
	return c.Redirect().To(h.oauthConfig.AuthCodeURL(state))
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GoogleCallback godoc
// @Summary Google OAuth Callback
// @Description Callback URL for Google OAuth2
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g., missing code)"
// @Failure 401 {object} map[string]interface{} "Unauthorized (failed to exchange token)"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/auth/google/callback [get]
func (h *Handler) GoogleCallback(c fiber.Ctx) error {
	stateParam := c.Query("state")
	cookieState := c.Cookies("oauth_state")
	c.ClearCookie("oauth_state")
	if stateParam == "" || cookieState == "" || stateParam != cookieState {
		return c.Redirect().To(h.frontendRedirectURL("error", "invalid_state"))
	}

	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Authorization code is missing from the callback request."})
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Failed to exchange authorization code for a Google token."})
	}

	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get(h.userInfoURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve user information from Google."})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Google responded with an error when fetching user information."})
	}

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process the user information received from Google."})
	}

	_, jwtToken, err := h.authUC.HandleGoogleCallback(c.Context(), userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		// ถ้ามีคำว่า unauthorized (เช่น หา email ไม่เจอในระบบ) ให้เด้งกลับไปพร้อม error
		if strings.Contains(err.Error(), "unauthorized") {
			return c.Redirect().To(h.frontendRedirectURL("error", "unauthorized_email"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Redirect().To(h.frontendRedirectURL("token", jwtToken))
}

func (h *Handler) frontendRedirectURL(key, value string) string {
	base := h.frontendURL
	if base == "" {
		base = "http://localhost:4200"
	}

	redirectURL, err := url.Parse(base + "/admin/login")
	if err != nil {
		return "http://localhost:4200/admin/login?" + url.Values{key: []string{value}}.Encode()
	}

	query := redirectURL.Query()
	query.Set(key, value)
	redirectURL.RawQuery = query.Encode()

	return redirectURL.String()
}

// Logout godoc
// @Summary Logout User
// @Description Logout the current user
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "logout successful",
	})
}
