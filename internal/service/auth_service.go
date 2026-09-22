package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"agentsight/internal/config"
	"agentsight/internal/models"
	"agentsight/internal/repository"
)

type AuthService interface {
	GetAuthURL(state string) string
	HandleCallback(ctx context.Context, code string) (*models.User, string, error)
	ValidateSession(ctx context.Context, token string) (*models.User, error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	httpClient  *http.Client
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository, sessionRepo repository.SessionRepository) AuthService {
	return &authService{
		cfg:         cfg,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *authService) GetAuthURL(state string) string {
	redirectURI := fmt.Sprintf("%s/auth/github/callback", strings.TrimSuffix(s.cfg.BaseURL, "/"))
	params := url.Values{}
	params.Set("client_id", s.cfg.GitHubClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "read:user,user:email")
	params.Set("state", state)
	return fmt.Sprintf("https://github.com/login/oauth/authorize?%s", params.Encode())
}

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type githubUserResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

type githubEmailResponse struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *authService) HandleCallback(ctx context.Context, code string) (*models.User, string, error) {
	if code == "" {
		return nil, "", errors.New("empty authorization code")
	}

	redirectURI := fmt.Sprintf("%s/auth/github/callback", strings.TrimSuffix(s.cfg.BaseURL, "/"))
	data := url.Values{}
	data.Set("client_id", s.cfg.GitHubClientID)
	data.Set("client_secret", s.cfg.GitHubClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("failed to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("oauth token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp githubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode oauth response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, "", fmt.Errorf("oauth error: %s (%s)", tokenResp.Error, tokenResp.ErrorDesc)
	}
	if tokenResp.AccessToken == "" {
		return nil, "", errors.New("missing access token from github")
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build user request: %w", err)
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("Accept", "application/json")
	userReq.Header.Set("User-Agent", "AgentSight-Server")

	userResp, err := s.httpClient.Do(userReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch github profile: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("github api returned status %d", userResp.StatusCode)
	}

	var ghUser githubUserResponse
	if err := json.NewDecoder(userResp.Body).Decode(&ghUser); err != nil {
		return nil, "", fmt.Errorf("failed to parse github profile: %w", err)
	}

	email := ghUser.Email
	if email == "" {
		emailReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
		if err == nil {
			emailReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
			emailReq.Header.Set("Accept", "application/json")
			emailReq.Header.Set("User-Agent", "AgentSight-Server")
			if emailResp, err := s.httpClient.Do(emailReq); err == nil {
				defer emailResp.Body.Close()
				var emails []githubEmailResponse
				if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
					for _, e := range emails {
						if e.Primary && e.Verified {
							email = e.Email
							break
						}
					}
					if email == "" && len(emails) > 0 {
						email = emails[0].Email
					}
				}
			}
		}
	}

	username := ghUser.Login
	if username == "" {
		username = ghUser.Name
	}
	if username == "" {
		username = fmt.Sprintf("gh-%d", ghUser.ID)
	}

	user := &models.User{
		GitHubID:  strconv.FormatInt(ghUser.ID, 10),
		Username:  username,
		AvatarURL: ghUser.AvatarURL,
		Email:     email,
	}

	persistedUser, err := s.userRepo.UpsertByGitHub(ctx, user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to persist user: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate session token: %w", err)
	}
	sessionToken := hex.EncodeToString(tokenBytes)

	session := &models.Session{
		Token:     sessionToken,
		UserID:    persistedUser.ID,
		ExpiresAt: time.Now().Add(14 * 24 * time.Hour),
	}
	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to store session: %w", err)
	}

	return persistedUser, sessionToken, nil
}

func (s *authService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("empty session token")
	}

	session, err := s.sessionRepo.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		_ = s.sessionRepo.DeleteSession(ctx, token)
		return nil, errors.New("session expired")
	}

	return s.userRepo.GetByID(ctx, session.UserID)
}

func (s *authService) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	return s.sessionRepo.DeleteSession(ctx, token)
}
