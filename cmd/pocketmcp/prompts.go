package pocketmcp

import (
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const defaultPocketBaseURL = "http://localhost:8090"

type pocketBaseCredentials struct {
	URL      string
	Email    string
	Password string
}

var setupClientOptions = []string{
	"claude-code",
	"claude-desktop",
	"codex",
	"cursor",
	"gemini",
	"opencode",
	"vscode",
	"windsurf",
}

func promptPocketBaseCredentials(cmd *cobra.Command, url string, email string, password string) (pocketBaseCredentials, error) {
	_ = cmd

	creds := pocketBaseCredentials{
		URL:      strings.TrimSpace(url),
		Email:    strings.TrimSpace(email),
		Password: strings.TrimSpace(password),
	}

	askOpts := []survey.AskOpt{
		survey.WithStdio(os.Stdin, os.Stdout, os.Stderr),
	}

	if creds.URL == "" {
		if err := survey.AskOne(
			&survey.Input{
				Message: "PocketBase URL:",
				Default: firstNonEmpty(os.Getenv("POCKETBASE_URL"), defaultPocketBaseURL),
			},
			&creds.URL,
			append(askOpts, survey.WithValidator(survey.Required))...,
		); err != nil {
			return pocketBaseCredentials{}, err
		}
	}

	if creds.Email == "" {
		if err := survey.AskOne(
			&survey.Input{
				Message: "PocketBase user/email:",
				Default: strings.TrimSpace(os.Getenv("POCKETBASE_EMAIL")),
			},
			&creds.Email,
			append(askOpts, survey.WithValidator(survey.Required))...,
		); err != nil {
			return pocketBaseCredentials{}, err
		}
	}

	if creds.Password == "" {
		if err := survey.AskOne(
			&survey.Password{
				Message: "PocketBase password:",
			},
			&creds.Password,
			append(askOpts, survey.WithValidator(survey.Required))...,
		); err != nil {
			return pocketBaseCredentials{}, err
		}
	}

	return creds, nil
}

func canPrompt() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func promptSetupClients(current string) ([]string, error) {
	defaults := parseClientSelection(current)
	if len(defaults) == 0 {
		defaults = []string{"claude-code", "codex", "cursor", "gemini", "opencode", "vscode", "windsurf"}
	}

	selected := append([]string(nil), defaults...)
	err := survey.AskOne(
		&survey.MultiSelect{
			Message: "AI clients to configure:",
			Options: setupClientOptions,
			Default: defaults,
		},
		&selected,
		survey.WithStdio(os.Stdin, os.Stdout, os.Stderr),
		survey.WithValidator(func(ans any) error {
			values, ok := ans.([]string)
			if !ok || len(values) == 0 {
				return survey.Required(ans)
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return selected, nil
}

func parseClientSelection(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	clients := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if value == "all" {
			return []string{"claude-code", "codex", "cursor", "gemini", "opencode", "vscode", "windsurf"}
		}
		clients = append(clients, value)
	}
	return clients
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
