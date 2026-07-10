package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	JWTSecret    string
	DashboardURL string
}

func HandleGitHubLogin(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=read:org", url.QueryEscape(cfg.ClientID))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

func HandleGitHubCallback(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code parameter", http.StatusBadRequest)
			return
		}

		// Exchange code for access token
		tokenURL := "https://github.com/login/oauth/access_token"
		tokenData := url.Values{
			"client_id":     {cfg.ClientID},
			"client_secret": {cfg.ClientSecret},
			"code":          {code},
		}

		req, err := http.NewRequest("POST", tokenURL, nil)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		req.URL.RawQuery = tokenData.Encode()
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "failed to exchange code for token", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenResp struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			http.Error(w, "failed to parse token response", http.StatusInternalServerError)
			return
		}

		if tokenResp.Error != "" || tokenResp.AccessToken == "" {
			http.Error(w, "failed to authenticate", http.StatusUnauthorized)
			return
		}

		// Fetch user orgs
		orgsReq, err := http.NewRequest("GET", "https://api.github.com/user/orgs", nil)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		orgsReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		orgsReq.Header.Set("Accept", "application/vnd.github+json")
		orgsReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		orgsResp, err := client.Do(orgsReq)
		if err != nil {
			http.Error(w, "failed to fetch user orgs", http.StatusInternalServerError)
			return
		}
		defer orgsResp.Body.Close()

		if orgsResp.StatusCode != http.StatusOK {
			http.Error(w, "failed to fetch user orgs", http.StatusInternalServerError)
			return
		}

		var orgsData []struct {
			Login string `json:"login"`
		}
		if err := json.NewDecoder(orgsResp.Body).Decode(&orgsData); err != nil {
			http.Error(w, "failed to parse orgs response", http.StatusInternalServerError)
			return
		}

		orgs := make([]string, 0, len(orgsData))
		for _, org := range orgsData {
			orgs = append(orgs, org.Login)
		}

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"orgs": orgs,
			"exp":  time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		// Redirect to Dashboard
		dashboardURL, err := url.Parse(cfg.DashboardURL)
		if err != nil {
			http.Error(w, "invalid dashboard url configured", http.StatusInternalServerError)
			return
		}
		q := dashboardURL.Query()
		q.Set("token", tokenString)
		dashboardURL.RawQuery = q.Encode()

		http.Redirect(w, r, dashboardURL.String(), http.StatusTemporaryRedirect)
	}
}
