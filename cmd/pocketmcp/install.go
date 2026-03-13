package pocketmcp

import (
	"fmt"
	"strings"

	"github.com/mreyeswilson/pocketmcp/internal/config"
	installer "github.com/mreyeswilson/pocketmcp/internal/install"
	"github.com/spf13/cobra"
)

type installOptions struct {
	Client    string
	Uninstall bool
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

func newSetupCmd() *cobra.Command {
	opts := installOptions{}

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Install MCP client configuration",
		Long:  "Write or remove MCP client configuration pointing at the global pocketmcp binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd, &opts)
		},
	}

	cmd.Flags().StringVar(&opts.Client, "client", "all", "Target client: all|claude-code|claude-desktop|codex|cursor|gemini|opencode|vscode|windsurf")
	cmd.Flags().BoolVar(&opts.Uninstall, "uninstall", false, "Remove existing entry instead of installing")
	cmd.Flags().StringVar(&opts.URL, "url", "", "PocketBase URL")
	cmd.Flags().StringVar(&opts.Email, "email", "", "PocketBase user email")
	cmd.Flags().StringVar(&opts.Email, "user", "", "Alias for --email")
	cmd.Flags().StringVar(&opts.Password, "password", "", "PocketBase password")
	cmd.Flags().IntVar(&opts.TimeoutMS, "timeout-ms", 0, "Request timeout in milliseconds")

	return cmd
}

func newUninstallCmd() *cobra.Command {
	opts := installOptions{Uninstall: true}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove MCP client configuration",
		Long:  "Remove MCP client configuration that points at the global pocketmcp binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd, &opts)
		},
	}

	cmd.Flags().StringVar(&opts.Client, "client", "all", "Target client: all|claude-code|claude-desktop|codex|cursor|gemini|opencode|vscode|windsurf")
	cmd.Flags().IntVar(&opts.TimeoutMS, "timeout-ms", 0, "Request timeout in milliseconds")

	return cmd
}

func runSetup(cmd *cobra.Command, opts *installOptions) error {
	selectedClients := parseClientSelection(opts.Client)

	if !opts.Uninstall {
		creds, err := promptPocketBaseCredentials(cmd, opts.URL, opts.Email, opts.Password)
		if err != nil {
			return err
		}
		opts.URL = creds.URL
		opts.Email = creds.Email
		opts.Password = creds.Password
	}
	if canPrompt() {
		clients, err := promptSetupClients(opts.Client)
		if err != nil {
			return err
		}
		selectedClients = clients
	}

	resolved, err := config.ResolveInstallConfig(config.InstallConfigInput{
		Client:    opts.Client,
		Clients:   selectedClients,
		Uninstall: opts.Uninstall,
		URL:       opts.URL,
		Email:     opts.Email,
		Password:  opts.Password,
		TimeoutMS: opts.TimeoutMS,
	})
	if err != nil {
		return err
	}

	results, err := installer.Apply(resolved)
	if err != nil {
		return err
	}

	action := "installed"
	if resolved.Uninstall {
		action = "removed"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s %s: client=%s binary=%s timeout_ms=%d\n",
		cmd.Name(),
		action,
		resolved.Client,
		valueOrDefault(installer.GlobalMCPBinary, "<auto>"),
		resolved.TimeoutMS,
	)
	if !resolved.Uninstall {
		fmt.Fprintf(cmd.OutOrStdout(), "setup auth: url=%s email=%s password=%s\n",
			resolved.URL,
			resolved.Email,
			config.RedactedPassword(resolved.Password),
		)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "targets: %s\n", strings.Join(resolved.TargetClients(), ", "))
	for _, result := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s -> %s\n", result.Action, result.Client, result.Path)
	}
	return nil
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
