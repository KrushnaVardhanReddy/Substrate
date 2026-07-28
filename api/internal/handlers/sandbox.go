package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

type SandboxHandler struct {
	Store     ports.Store
	JWTSecret string
}

type SandboxTokenRequest struct {
	Org  string `json:"org"`
	Repo string `json:"repo"`
}

type SandboxProxyRequest struct {
	Org     string            `json:"org"`
	Repo    string            `json:"repo"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type SandboxProxyResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func (h *SandboxHandler) GenerateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req SandboxTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Org == "" || req.Repo == "" {
		http.Error(w, "org and repo are required", http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256([]byte(h.JWTSecret))
	key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
	if err != nil {
		http.Error(w, "internal server error: invalid key", http.StatusInternalServerError)
		return
	}

	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(15 * time.Minute))
	err = token.Set("org", req.Org)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = token.Set("repo", req.Repo)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = token.Set("sandbox", true)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tokenString := token.V4Encrypt(key, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (h *SandboxHandler) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req SandboxProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
		return
	}
	tokenStr := parts[1]

	hash := sha256.Sum256([]byte(h.JWTSecret))
	key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
	if err != nil {
		http.Error(w, "internal server error: invalid key", http.StatusInternalServerError)
		return
	}

	parser := paseto.NewParser()
	parsedToken, err := parser.ParseV4Local(key, tokenStr, nil)
	if err != nil {
		http.Error(w, "invalid or expired sandbox token", http.StatusUnauthorized)
		return
	}

	var isSandbox bool
	if err := parsedToken.Get("sandbox", &isSandbox); err != nil || !isSandbox {
		http.Error(w, "invalid token scope", http.StatusForbidden)
		return
	}

	var tokenOrg, tokenRepo string
	if err := parsedToken.Get("org", &tokenOrg); err != nil || tokenOrg != req.Org {
		http.Error(w, "invalid token org", http.StatusForbidden)
		return
	}
	if err := parsedToken.Get("repo", &tokenRepo); err != nil || tokenRepo != req.Repo {
		http.Error(w, "invalid token repo", http.StatusForbidden)
		return
	}

	// Retrieve base_url
	orgID, err := h.Store.GetOrgIDByName(r.Context(), req.Org)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	queries := sqlcgen.New(h.Store.Pool())
	baseURLText, err := queries.GetRepoBaseURL(r.Context(), sqlcgen.GetRepoBaseURLParams{
		OrgID: pgtype.UUID{Bytes: orgID, Valid: true},
		Name:  req.Repo,
	})
	if err != nil || !baseURLText.Valid || baseURLText.String == "" {
		http.Error(w, "repository has no base URL configured", http.StatusBadRequest)
		return
	}
	baseURL := baseURLText.String

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		http.Error(w, "invalid repository base URL", http.StatusInternalServerError)
		return
	}

	// SSRF Prevention
	host := parsedBaseURL.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		http.Error(w, "failed to resolve host", http.StatusBadRequest)
		return
	}
	for _, ip := range ips {
		// Explicitly check for 169.254.x.x link-local and other private ranges
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.String() == "169.254.169.254" {
			http.Error(w, "forbidden destination (private IP range)", http.StatusForbidden)
			return
		}
	}

	// Construct Proxy Target URL
	targetURL := baseURL
	if !strings.HasSuffix(targetURL, "/") && !strings.HasPrefix(req.Path, "/") {
		targetURL += "/"
	}
	targetURL += strings.TrimPrefix(req.Path, "/")

	// Create Request
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	outReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bodyReader)
	if err != nil {
		http.Error(w, "failed to create proxy request", http.StatusInternalServerError)
		return
	}

	for k, v := range req.Headers {
		outReq.Header.Set(k, v)
	}
	outReq.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(outReq)
	if err != nil {
		http.Error(w, "failed to execute proxy request: " + err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response body", http.StatusInternalServerError)
		return
	}

	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	// Log to sandbox_audit_log
	// UserID needs to be extracted from the authenticated user. We assume a dummy user for now or read from context
	userID := "sandbox-user"
	queries.InsertSandboxAuditLog(context.Background(), sqlcgen.InsertSandboxAuditLogParams{
		Org:        req.Org,
		Repo:       req.Repo,
		Method:     req.Method,
		Path:       req.Path,
		StatusCode: pgtype.Int4{Int32: int32(resp.StatusCode), Valid: true},
		UserID:     userID,
	})

	proxyResp := SandboxProxyResponse{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(respBody),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proxyResp)
}
