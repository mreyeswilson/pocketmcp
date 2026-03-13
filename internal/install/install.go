package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mreyeswilson/pocketmcp/internal/config"
	"github.com/pelletier/go-toml/v2"
)

const (
	ServerName      = "pocketbase-admin"
	GlobalMCPBinary = "pocketmcp"
)

type Result struct {
	Client string
	Path   string
	Action string
}

type clientSpec struct {
	Name string
	Path string
	Kind string
	Key  string
}

func Apply(cfg config.InstallConfig) ([]Result, error) {
	results := make([]Result, 0, len(cfg.TargetClients()))
	for _, client := range cfg.TargetClients() {
		spec, err := resolveClientSpec(client)
		if err != nil {
			return nil, err
		}

		if err := os.MkdirAll(filepath.Dir(spec.Path), 0o755); err != nil {
			return nil, fmt.Errorf("create config directory for %s: %w", client, err)
		}

		action := "updated"
		if cfg.Uninstall {
			if err := removeConfig(spec); err != nil {
				return nil, fmt.Errorf("%s uninstall failed: %w", client, err)
			}
			action = "removed"
		} else {
			if err := upsertConfig(spec, cfg); err != nil {
				return nil, fmt.Errorf("%s install failed: %w", client, err)
			}
			action = "updated"
		}

		results = append(results, Result{
			Client: spec.Name,
			Path:   spec.Path,
			Action: action,
		})
	}

	return results, nil
}

func resolveClientSpec(client string) (clientSpec, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return clientSpec{}, fmt.Errorf("resolve user home: %w", err)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return clientSpec{}, fmt.Errorf("resolve user config dir: %w", err)
	}

	switch client {
	case "claude-code":
		return clientSpec{Name: client, Path: filepath.Join(homeDir, ".claude.json"), Kind: "json-mcpServers", Key: "mcpServers"}, nil
	case "claude-desktop":
		return clientSpec{Name: client, Path: claudeDesktopPath(homeDir, configDir), Kind: "json-mcpServers", Key: "mcpServers"}, nil
	case "codex":
		return clientSpec{Name: client, Path: filepath.Join(homeDir, ".codex", "config.toml"), Kind: "toml-mcpServers", Key: "mcp_servers"}, nil
	case "cursor":
		return clientSpec{Name: client, Path: filepath.Join(homeDir, ".cursor", "mcp.json"), Kind: "json-mcpServers", Key: "mcpServers"}, nil
	case "gemini":
		return clientSpec{Name: client, Path: filepath.Join(homeDir, ".gemini", "settings.json"), Kind: "json-mcpServers", Key: "mcpServers"}, nil
	case "opencode":
		return clientSpec{Name: client, Path: filepath.Join(configDir, "opencode", "opencode.json"), Kind: "json-opencode", Key: "mcp"}, nil
	case "vscode":
		return clientSpec{Name: client, Path: filepath.Join(configDir, "Code", "User", "mcp.json"), Kind: "json-vscode", Key: "servers"}, nil
	case "windsurf":
		return clientSpec{Name: client, Path: filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json"), Kind: "json-mcpServers", Key: "mcpServers"}, nil
	default:
		return clientSpec{}, fmt.Errorf("unsupported client: %s", client)
	}
}

func claudeDesktopPath(homeDir string, configDir string) string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(configDir, "Claude", "claude_desktop_config.json")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(configDir, "Claude", "claude_desktop_config.json")
	}
}

func upsertConfig(spec clientSpec, cfg config.InstallConfig) error {
	switch spec.Kind {
	case "json-mcpServers", "json-vscode":
		entry := map[string]any{
			"command": GlobalMCPBinary,
			"args":    mcpArgs(cfg),
			"env":     map[string]string{},
		}
		return upsertJSONEntry(spec.Path, spec.Key, entry)
	case "json-opencode":
		entry := map[string]any{
			"type":        "local",
			"command":     append([]string{GlobalMCPBinary}, mcpArgs(cfg)...),
			"enabled":     true,
			"environment": map[string]string{},
			"timeout":     cfg.TimeoutMS,
		}
		return upsertJSONEntry(spec.Path, spec.Key, entry)
	case "toml-mcpServers":
		entry := map[string]any{
			"command": GlobalMCPBinary,
			"args":    mcpArgs(cfg),
			"env":     map[string]string{},
		}
		return upsertTOMLEntry(spec.Path, spec.Key, entry)
	default:
		return fmt.Errorf("unsupported config kind: %s", spec.Kind)
	}
}

func removeConfig(spec clientSpec) error {
	switch spec.Kind {
	case "json-mcpServers", "json-vscode", "json-opencode":
		return removeJSONEntry(spec.Path, spec.Key)
	case "toml-mcpServers":
		return removeTOMLEntry(spec.Path, spec.Key)
	default:
		return fmt.Errorf("unsupported config kind: %s", spec.Kind)
	}
}

func mcpArgs(cfg config.InstallConfig) []string {
	return []string{
		"mcp",
		"--url", cfg.URL,
		"--email", cfg.Email,
		"--password", cfg.Password,
		"--timeout-ms", strconv.Itoa(cfg.TimeoutMS),
	}
}

func upsertJSONEntry(path string, topKey string, entry map[string]any) error {
	doc, err := loadJSONDocument(path)
	if err != nil {
		return err
	}

	target, err := nestedObject(doc, topKey, true)
	if err != nil {
		return err
	}
	target[ServerName] = entry

	return saveJSONDocument(path, doc)
}

func removeJSONEntry(path string, topKey string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat JSON config: %w", err)
	}

	doc, err := loadJSONDocument(path)
	if err != nil {
		return err
	}

	target, err := nestedObject(doc, topKey, false)
	if err != nil {
		return err
	}
	if target == nil {
		return nil
	}

	delete(target, ServerName)
	if len(target) == 0 {
		delete(doc, topKey)
	}

	return saveJSONDocument(path, doc)
}

func upsertTOMLEntry(path string, topKey string, entry map[string]any) error {
	doc, err := loadTOMLDocument(path)
	if err != nil {
		return err
	}

	target, err := nestedObject(doc, topKey, true)
	if err != nil {
		return err
	}
	target[ServerName] = entry

	data, err := toml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal TOML: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func removeTOMLEntry(path string, topKey string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat TOML config: %w", err)
	}

	doc, err := loadTOMLDocument(path)
	if err != nil {
		return err
	}

	target, err := nestedObject(doc, topKey, false)
	if err != nil {
		return err
	}
	if target == nil {
		return nil
	}

	delete(target, ServerName)
	if len(target) == 0 {
		delete(doc, topKey)
	}

	data, err := toml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal TOML: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func loadJSONDocument(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read JSON config: %w", err)
	}
	if strings.TrimSpace(string(raw)) == "" {
		return map[string]any{}, nil
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse JSON config: %w", err)
	}
	if doc == nil {
		doc = map[string]any{}
	}
	return doc, nil
}

func saveJSONDocument(path string, doc map[string]any) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON config: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func loadTOMLDocument(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read TOML config: %w", err)
	}
	if strings.TrimSpace(string(raw)) == "" {
		return map[string]any{}, nil
	}

	var doc map[string]any
	if err := toml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse TOML config: %w", err)
	}
	if doc == nil {
		doc = map[string]any{}
	}
	return doc, nil
}

func nestedObject(doc map[string]any, key string, create bool) (map[string]any, error) {
	if value, ok := doc[key]; ok {
		obj, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s must be an object", key)
		}
		return obj, nil
	}
	if !create {
		return nil, nil
	}

	obj := map[string]any{}
	doc[key] = obj
	return obj, nil
}
