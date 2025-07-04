package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"excalidraw-complete/core"
	"excalidraw-complete/stores"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthHandler struct {
	store        stores.Store
	oauthConfig  *oauth2.Config
	baseURL      string
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type LoginResponse struct {
	User         *core.User `json:"user"`
	SessionToken string     `json:"session_token"`
}

func NewAuthHandler(store stores.Store) *AuthHandler {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	baseURL := os.Getenv("BASE_URL")
	
	if baseURL == "" {
		baseURL = "http://localhost:3002"
	}
	
	if clientID == "" || clientSecret == "" {
		logrus.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET environment variables must be set")
	}
	
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
	
	return &AuthHandler{
		store:       store,
		oauthConfig: oauthConfig,
		baseURL:     baseURL,
	}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	r.Get("/github", h.handleGitHubLogin)
	r.Get("/github/callback", h.handleGitHubCallback)
	r.Post("/logout", h.handleLogout)
	r.Get("/me", h.handleMe)
	
	return r
}

func (h *AuthHandler) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := generateRandomState()
	
	// 将state保存到cookie中以验证回调
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10分钟
	}
	http.SetCookie(w, cookie)
	
	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// 验证state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	
	// 清除state cookie
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}
	
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		logrus.WithError(err).Error("Failed to exchange token")
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	
	// 获取GitHub用户信息
	githubUser, err := h.getGitHubUser(token.AccessToken)
	if err != nil {
		logrus.WithError(err).Error("Failed to get GitHub user")
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	
	// 创建或更新用户
	user, err := h.createOrUpdateUser(r.Context(), githubUser)
	if err != nil {
		logrus.WithError(err).Error("Failed to create or update user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	
	// 创建会话
	session := &core.Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7天
	}
	
	err = h.store.GetSessionStore().CreateSession(r.Context(), session)
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	
	// 设置会话cookie
	sessionCookie := &http.Cookie{
		Name:     "session_token",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   604800, // 7天
	}
	http.SetCookie(w, sessionCookie)
	
	// 重定向到前端
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		// 删除会话
		h.store.GetSessionStore().DeleteSession(r.Context(), sessionCookie.Value)
	}
	
	// 清除会话cookie
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	
	render.JSON(w, r, map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	render.JSON(w, r, user)
}

func (h *AuthHandler) getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var githubUser GitHubUser
	err = json.Unmarshal(body, &githubUser)
	if err != nil {
		return nil, err
	}
	
	return &githubUser, nil
}

func (h *AuthHandler) createOrUpdateUser(ctx context.Context, githubUser *GitHubUser) (*core.User, error) {
	// 检查用户是否已存在
	existingUser, err := h.store.GetUserStore().GetUserByGitHubID(ctx, githubUser.ID)
	if err != nil {
		return nil, err
	}
	
	if existingUser != nil {
		// 更新现有用户
		existingUser.Username = githubUser.Login
		existingUser.Name = githubUser.Name
		existingUser.Email = githubUser.Email
		existingUser.AvatarURL = githubUser.AvatarURL
		
		err = h.store.GetUserStore().UpdateUser(ctx, existingUser)
		if err != nil {
			return nil, err
		}
		
		return existingUser, nil
	}
	
	// 创建新用户
	user := &core.User{
		GitHubID:  githubUser.ID,
		Username:  githubUser.Login,
		Name:      githubUser.Name,
		Email:     githubUser.Email,
		AvatarURL: githubUser.AvatarURL,
	}
	
	err = h.store.GetUserStore().CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}