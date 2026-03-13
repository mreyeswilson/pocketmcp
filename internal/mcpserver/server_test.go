package mcpserver

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mreyeswilson/pocketmcp/internal/pocketbase"
)

func TestServerRegistersExpectedTools(t *testing.T) {
	pb := pocketbase.NewClient("http://127.0.0.1:8090", "admin@example.com", "secret", time.Second)
	server := New(pb, "test")

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect returned error: %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect returned error: %v", err)
	}
	defer clientSession.Close()

	tools := map[string]bool{}
	for tool, err := range clientSession.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("Tools iterator returned error: %v", err)
		}
		tools[tool.Name] = true
	}

	expected := []string{
		"list_collections",
		"get_collection",
		"create_collection",
		"update_collection",
		"delete_collection",
		"list_records",
		"get_record",
		"create_record",
		"update_record",
		"delete_record",
		"get_settings",
		"update_settings",
	}
	for _, name := range expected {
		if !tools[name] {
			t.Fatalf("expected tool %q to be registered", name)
		}
	}
}
