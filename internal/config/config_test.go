package config

import (
	"os"
	"testing"
)

func TestResolveServeConfigUsesEnvFallback(t *testing.T) {
	t.Setenv("POCKETBASE_URL", "http://127.0.0.1:8090")
	t.Setenv("POCKETBASE_EMAIL", "admin@example.com")
	t.Setenv("POCKETBASE_PASSWORD", "secret")
	t.Setenv("REQUEST_TIMEOUT_MS", "1234")

	resolved, err := ResolveServeConfig(ServeConfigInput{})
	if err != nil {
		t.Fatalf("ResolveServeConfig returned error: %v", err)
	}

	if resolved.URL != "http://127.0.0.1:8090" {
		t.Fatalf("unexpected URL: %s", resolved.URL)
	}
	if resolved.Email != "admin@example.com" {
		t.Fatalf("unexpected email: %s", resolved.Email)
	}
	if resolved.Password != "secret" {
		t.Fatalf("unexpected password: %s", resolved.Password)
	}
	if resolved.TimeoutMS != 1234 {
		t.Fatalf("unexpected timeout: %d", resolved.TimeoutMS)
	}
}

func TestResolveInstallConfigForUninstallDoesNotRequireCredentials(t *testing.T) {
	resolved, err := ResolveInstallConfig(InstallConfigInput{Client: "all", Uninstall: true})
	if err != nil {
		t.Fatalf("ResolveInstallConfig returned error: %v", err)
	}

	if !resolved.Uninstall {
		t.Fatal("expected uninstall to be true")
	}
	if len(resolved.TargetClients()) != 7 {
		t.Fatalf("unexpected targets: %v", resolved.TargetClients())
	}
}

func TestResolveInstallConfigRejectsUnsupportedClient(t *testing.T) {
	_, err := ResolveInstallConfig(InstallConfigInput{Client: "unknown", Uninstall: true})
	if err == nil {
		t.Fatal("expected error for unsupported client")
	}
}

func TestResolveInstallConfigNormalizesClientName(t *testing.T) {
	resolved, err := ResolveInstallConfig(InstallConfigInput{Client: "Windsurf", Uninstall: true})
	if err != nil {
		t.Fatalf("ResolveInstallConfig returned error: %v", err)
	}

	if resolved.Client != "windsurf" {
		t.Fatalf("unexpected normalized client: %s", resolved.Client)
	}
}

func TestResolveInstallConfigSupportsCodex(t *testing.T) {
	resolved, err := ResolveInstallConfig(InstallConfigInput{Client: "Codex", Uninstall: true})
	if err != nil {
		t.Fatalf("ResolveInstallConfig returned error: %v", err)
	}

	if resolved.Client != "codex" {
		t.Fatalf("unexpected normalized client: %s", resolved.Client)
	}
}

func TestResolveInstallConfigSupportsMultipleClients(t *testing.T) {
	resolved, err := ResolveInstallConfig(InstallConfigInput{Clients: []string{"codex", "gemini"}, Uninstall: true})
	if err != nil {
		t.Fatalf("ResolveInstallConfig returned error: %v", err)
	}

	targets := resolved.TargetClients()
	if len(targets) != 2 || targets[0] != "codex" || targets[1] != "gemini" {
		t.Fatalf("unexpected targets: %v", targets)
	}
}

func TestResolveServeConfigRejectsMissingValues(t *testing.T) {
	os.Unsetenv("POCKETBASE_URL")
	os.Unsetenv("POCKETBASE_EMAIL")
	os.Unsetenv("POCKETBASE_PASSWORD")

	_, err := ResolveServeConfig(ServeConfigInput{})
	if err == nil {
		t.Fatal("expected missing configuration error")
	}
}

func TestResolveServeConfigRejectsInvalidTimeoutEnv(t *testing.T) {
	t.Setenv("POCKETBASE_URL", "http://127.0.0.1:8090")
	t.Setenv("POCKETBASE_EMAIL", "admin@example.com")
	t.Setenv("POCKETBASE_PASSWORD", "secret")
	t.Setenv("REQUEST_TIMEOUT_MS", "bad")

	_, err := ResolveServeConfig(ServeConfigInput{})
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
}
