package pocketbase

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientAuthenticatesAndListsCollections(t *testing.T) {
	var authCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authCalls.Add(1)
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"token": "secret-token",
				"record": map[string]any{
					"id": "superuser",
				},
			})
		case "/api/collections":
			if got := r.Header.Get("Authorization"); got != "secret-token" {
				t.Fatalf("unexpected authorization header: %q", got)
			}
			if got := r.URL.Query().Get("perPage"); got != "50" {
				t.Fatalf("unexpected perPage: %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{{"name": "users"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin@example.com", "secret", 2*time.Second)
	result, err := client.ListCollections(context.Background(), url.Values{"perPage": {"50"}})
	if err != nil {
		t.Fatalf("ListCollections returned error: %v", err)
	}

	items, ok := result["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected items payload: %#v", result["items"])
	}
	if authCalls.Load() != 1 {
		t.Fatalf("expected one auth call, got %d", authCalls.Load())
	}
}

func TestClientCreateRecordRetriesOnUnauthorized(t *testing.T) {
	var authenticated atomic.Bool
	var authCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authCalls.Add(1)
			authenticated.Store(true)
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "token-1"})
		case "/api/collections/users/records":
			if !authenticated.Load() || r.Header.Get("Authorization") == "" {
				http.Error(w, `{"status":401,"message":"unauthorized","data":{}}`, http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll returned error: %v", err)
			}
			if string(body) != `{"email":"user@example.com"}` {
				t.Fatalf("unexpected body: %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "record1"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin@example.com", "secret", 2*time.Second)
	result, err := client.CreateRecord(context.Background(), "users", map[string]any{"email": "user@example.com"}, nil)
	if err != nil {
		t.Fatalf("CreateRecord returned error: %v", err)
	}

	if result["id"] != "record1" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if authCalls.Load() != 1 {
		t.Fatalf("expected one auth call, got %d", authCalls.Load())
	}
}
