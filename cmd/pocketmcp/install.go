package pocketmcp

import (
	"fmt"
	"strings"

	"github.com/mreyeswilson/pocketmcp/internal/config"
	"github.com/spf13/cobra"
)

type installOptions struct {
	Client    string
	Uninstall bool
	Binary    string
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

func newSetupCmd() *cobra.Command {
	opts := installOptions{}

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Validate and prepare MCP client setup",
		Long:  "Validate local MCP client installation parameters before writing client configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.Uninstall {
				creds, err := promptPocketBaseCredentials(cmd, opts.URL, opts.Email, opts.Password)
				if err != nil {
					return err
				}
				opts.URL = creds.URL
				opts.Email = creds.Email
				opts.Password = creds.Password
			}

			resolved, err := config.ResolveInstallConfig(config.InstallConfigInput{
				Client:    opts.Client,
				Uninstall: opts.Uninstall,
				Binary:    opts.Binary,
				URL:       opts.URL,
				Email:     opts.Email,
				Password:  opts.Password,
				TimeoutMS: opts.TimeoutMS,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "setup config valid: client=%s uninstall=%t binary=%s timeout_ms=%d\n",
				resolved.Client,
				resolved.Uninstall,
				valueOrDefault(resolved.Binary, "<auto>"),
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
			fmt.Fprintln(cmd.OutOrStdout(), "next: implement client config path resolution and JSON patch/write logic")
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Client, "client", "all", "Target client: all|claude-desktop|cursor|vscode|windsurf|opencode")
	cmd.Flags().BoolVar(&opts.Uninstall, "uninstall", false, "Remove existing entry instead of installing")
	cmd.Flags().StringVar(&opts.Binary, "binary", "", "Path to compiled CLI binary")
	cmd.Flags().StringVar(&opts.URL, "url", "", "PocketBase URL")
	cmd.Flags().StringVar(&opts.Email, "email", "", "PocketBase user email")
	cmd.Flags().StringVar(&opts.Email, "user", "", "Alias for --email")
	cmd.Flags().StringVar(&opts.Password, "password", "", "PocketBase password")
	cmd.Flags().IntVar(&opts.TimeoutMS, "timeout-ms", 0, "Request timeout in milliseconds")

	return cmd
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
