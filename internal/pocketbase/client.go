package pocketbase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL    string
	email      string
	password   string
	httpClient *http.Client

	mu    sync.Mutex
	token string
}

type ErrorResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type AuthResponse struct {
	Token  string         `json:"token"`
	Record map[string]any `json:"record"`
}

func NewClient(baseURL string, email string, password string, timeout time.Duration) *Client {
	return &Client{
		baseURL:  strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		email:    strings.TrimSpace(email),
		password: password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) ListCollections(ctx context.Context, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, "/api/collections", params, nil)
}

func (c *Client) GetCollection(ctx context.Context, collection string, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, "/api/collections/"+url.PathEscape(collection), params, nil)
}

func (c *Client) CreateCollection(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPost, "/api/collections", nil, body)
}

func (c *Client) UpdateCollection(ctx context.Context, collection string, body map[string]any) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPatch, "/api/collections/"+url.PathEscape(collection), nil, body)
}

func (c *Client) DeleteCollection(ctx context.Context, collection string) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodDelete, "/api/collections/"+url.PathEscape(collection), nil, nil)
}

func (c *Client) ListRecords(ctx context.Context, collection string, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, "/api/collections/"+url.PathEscape(collection)+"/records", params, nil)
}

func (c *Client) GetRecord(ctx context.Context, collection string, recordID string, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, "/api/collections/"+url.PathEscape(collection)+"/records/"+url.PathEscape(recordID), params, nil)
}

func (c *Client) CreateRecord(ctx context.Context, collection string, body map[string]any, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPost, "/api/collections/"+url.PathEscape(collection)+"/records", params, body)
}

func (c *Client) UpdateRecord(ctx context.Context, collection string, recordID string, body map[string]any, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPatch, "/api/collections/"+url.PathEscape(collection)+"/records/"+url.PathEscape(recordID), params, body)
}

func (c *Client) DeleteRecord(ctx context.Context, collection string, recordID string) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodDelete, "/api/collections/"+url.PathEscape(collection)+"/records/"+url.PathEscape(recordID), nil, nil)
}

func (c *Client) GetSettings(ctx context.Context, params url.Values) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, "/api/settings", params, nil)
}

func (c *Client) UpdateSettings(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPatch, "/api/settings", nil, body)
}

func (c *Client) requestJSON(ctx context.Context, method string, path string, params url.Values, body any) (map[string]any, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, err
	}

	data, status, err := c.doRequest(ctx, method, path, params, body, true)
	if err != nil && status == http.StatusUnauthorized {
		c.clearToken()
		if authErr := c.ensureAuthenticated(ctx); authErr != nil {
			return nil, authErr
		}
		data, _, err = c.doRequest(ctx, method, path, params, body, true)
	}
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 || status == http.StatusNoContent {
		return map[string]any{"ok": true}, nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("decode PocketBase response: %w", err)
	}
	if decoded == nil {
		decoded = map[string]any{"ok": true}
	}
	return decoded, nil
}

func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.mu.Lock()
	if c.token != "" {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	body := map[string]string{
		"identity": c.email,
		"password": c.password,
	}
	data, _, err := c.doRequest(ctx, http.MethodPost, "/api/collections/_superusers/auth-with-password", nil, body, false)
	if err != nil {
		return fmt.Errorf("authenticate superuser: %w", err)
	}

	var auth AuthResponse
	if err := json.Unmarshal(data, &auth); err != nil {
		return fmt.Errorf("decode auth response: %w", err)
	}
	if strings.TrimSpace(auth.Token) == "" {
		return fmt.Errorf("authenticate superuser: empty auth token")
	}

	c.mu.Lock()
	c.token = auth.Token
	c.mu.Unlock()
	return nil
}

func (c *Client) doRequest(ctx context.Context, method string, path string, params url.Values, body any, includeAuth bool) ([]byte, int, error) {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("encode request body: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}

	endpoint := c.baseURL + path
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if includeAuth {
		c.mu.Lock()
		token := c.token
		c.mu.Unlock()
		if token != "" {
			req.Header.Set("Authorization", token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var pbErr ErrorResponse
		if json.Unmarshal(data, &pbErr) == nil && pbErr.Message != "" {
			return nil, resp.StatusCode, fmt.Errorf("%s (status %d)", pbErr.Message, resp.StatusCode)
		}
		return nil, resp.StatusCode, fmt.Errorf("pocketbase request failed with status %d", resp.StatusCode)
	}

	return data, resp.StatusCode, nil
}

func (c *Client) clearToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
}
