package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mreyeswilson/pocketmcp/internal/config"
)

func TestUpsertJSONEntryCreatesPocketBaseServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mcp.json")

	err := upsertJSONEntry(path, "mcpServers", map[string]any{
		"command": GlobalMCPBinary,
		"args":    []string{"mcp", "--url", "http://localhost:8090"},
		"env":     map[string]string{},
	})
	if err != nil {
		t.Fatalf("upsertJSONEntry returned error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	text := string(raw)
	if !strings.Contains(text, "\"pocketbase-admin\"") {
		t.Fatalf("expected pocketbase-admin entry in %s", text)
	}
	if !strings.Contains(text, "\"command\": \"pocketmcp\"") {
		t.Fatalf("expected global pocketmcp command in %s", text)
	}
}

func TestRemoveJSONEntryDeletesOnlyPocketBaseServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mcp.json")
	initial := `{
  "mcpServers": {
    "pocketbase-admin": {"command":"pocketmcp","args":["mcp"]},
    "other": {"command":"npx","args":["-y","foo"]}
  }
}
`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := removeJSONEntry(path, "mcpServers"); err != nil {
		t.Fatalf("removeJSONEntry returned error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	text := string(raw)
	if strings.Contains(text, "pocketbase-admin") {
		t.Fatalf("expected pocketbase-admin entry removed from %s", text)
	}
	if !strings.Contains(text, "\"other\"") {
		t.Fatalf("expected other entry to remain in %s", text)
	}
}

func TestUpsertTOMLEntryCreatesCodexServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	err := upsertTOMLEntry(path, "mcp_servers", map[string]any{
		"command": GlobalMCPBinary,
		"args":    mcpArgs(config.InstallConfig{URL: "http://localhost:8090", Email: "admin@example.com", Password: "secret", TimeoutMS: 15000}),
		"env":     map[string]string{},
	})
	if err != nil {
		t.Fatalf("upsertTOMLEntry returned error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	text := string(raw)
	if !strings.Contains(text, "[mcp_servers.pocketbase-admin]") {
		t.Fatalf("expected codex mcp server section in %s", text)
	}
	if !strings.Contains(text, "command = 'pocketmcp'") && !strings.Contains(text, "command = \"pocketmcp\"") {
		t.Fatalf("expected global pocketmcp command in %s", text)
	}
}
