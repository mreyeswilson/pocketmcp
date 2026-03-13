package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const DefaultTimeoutMS = 15000

var supportedClients = map[string]struct{}{
	"all":            {},
	"claude-code":    {},
	"claude-desktop": {},
	"codex":          {},
	"cursor":         {},
	"gemini":         {},
	"opencode":       {},
	"vscode":         {},
	"windsurf":       {},
}

type ServeConfigInput struct {
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

type ServeConfig struct {
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

type InstallConfigInput struct {
	Client    string
	Clients   []string
	Uninstall bool
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

type InstallConfig struct {
	Client    string
	Clients   []string
	Uninstall bool
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

func ResolveServeConfig(input ServeConfigInput) (ServeConfig, error) {
	url := firstNonEmpty(input.URL, os.Getenv("POCKETBASE_URL"))
	email := firstNonEmpty(input.Email, os.Getenv("POCKETBASE_EMAIL"))
	password := firstNonEmpty(input.Password, os.Getenv("POCKETBASE_PASSWORD"))
	timeoutMS, err := resolveTimeoutMS(input.TimeoutMS)
	if err != nil {
		return ServeConfig{}, err
	}

	missing := make([]string, 0, 3)
	if url == "" {
		missing = append(missing, "--url or POCKETBASE_URL")
	}
	if email == "" {
		missing = append(missing, "--email/--user or POCKETBASE_EMAIL")
	}
	if password == "" {
		missing = append(missing, "--password or POCKETBASE_PASSWORD")
	}
	if len(missing) > 0 {
		return ServeConfig{}, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return ServeConfig{
		URL:       url,
		Email:     email,
		Password:  password,
		TimeoutMS: timeoutMS,
	}, nil
}

func ResolveInstallConfig(input InstallConfigInput) (InstallConfig, error) {
	clients, err := resolveClients(input.Clients, input.Client)
	if err != nil {
		return InstallConfig{}, err
	}

	timeoutMS, err := resolveTimeoutMS(input.TimeoutMS)
	if err != nil {
		return InstallConfig{}, err
	}

	resolved := InstallConfig{
		Client:    strings.Join(clients, ","),
		Clients:   clients,
		Uninstall: input.Uninstall,
		TimeoutMS: timeoutMS,
	}

	if resolved.Uninstall {
		return resolved, nil
	}

	serveConfig, err := ResolveServeConfig(ServeConfigInput{
		URL:       input.URL,
		Email:     input.Email,
		Password:  input.Password,
		TimeoutMS: input.TimeoutMS,
	})
	if err != nil {
		return InstallConfig{}, err
	}

	resolved.URL = serveConfig.URL
	resolved.Email = serveConfig.Email
	resolved.Password = serveConfig.Password
	resolved.TimeoutMS = serveConfig.TimeoutMS

	return resolved, nil
}

func (c InstallConfig) TargetClients() []string {
	if len(c.Clients) > 0 {
		return append([]string(nil), c.Clients...)
	}
	if c.Client == "all" {
		return defaultInstallClients()
	}
	return []string{c.Client}
}

func resolveClients(explicitClients []string, rawClient string) ([]string, error) {
	candidates := explicitClients
	if len(candidates) == 0 {
		candidates = strings.Split(firstNonEmpty(rawClient, "all"), ",")
	}

	seen := map[string]struct{}{}
	resolved := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		client := normalizeClient(candidate)
		if client == "" {
			continue
		}
		if client == "all" {
			for _, defaultClient := range defaultInstallClients() {
				if _, ok := seen[defaultClient]; ok {
					continue
				}
				seen[defaultClient] = struct{}{}
				resolved = append(resolved, defaultClient)
			}
			continue
		}
		if err := validateClient(client); err != nil {
			return nil, err
		}
		if _, ok := seen[client]; ok {
			continue
		}
		seen[client] = struct{}{}
		resolved = append(resolved, client)
	}

	if len(resolved) == 0 {
		return nil, errors.New("at least one client must be selected")
	}
	return resolved, nil
}

func defaultInstallClients() []string {
	return []string{"claude-code", "codex", "cursor", "gemini", "opencode", "vscode", "windsurf"}
}

func RedactedPassword(password string) string {
	if strings.TrimSpace(password) == "" {
		return "<empty>"
	}
	return "***"
}

func resolveTimeoutMS(flagValue int) (int, error) {
	if flagValue > 0 {
		return flagValue, nil
	}

	envValue := strings.TrimSpace(os.Getenv("REQUEST_TIMEOUT_MS"))
	if envValue == "" {
		return DefaultTimeoutMS, nil
	}

	parsed, err := parsePositiveInt(envValue)
	if err != nil {
		return 0, fmt.Errorf("REQUEST_TIMEOUT_MS must be a positive integer")
	}
	return parsed, nil
}

func validateClient(value string) error {
	normalized := normalizeClient(value)
	if normalized == "" {
		return errors.New("client must be a non-empty string")
	}
	if _, ok := supportedClients[normalized]; !ok {
		return fmt.Errorf("unsupported client: %s", value)
	}
	return nil
}

func normalizeClient(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func parsePositiveInt(raw string) (int, error) {
	var value int
	_, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &value)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("value must be a positive integer: %q", raw)
	}
	return value, nil
}
