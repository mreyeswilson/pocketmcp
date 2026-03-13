package pocketmcp

import (
	"fmt"

	"github.com/mreyeswilson/pocketmcp/internal/config"
	"github.com/spf13/cobra"
)

type serveOptions struct {
	URL       string
	Email     string
	Password  string
	TimeoutMS int
}

func newMCPCmd() *cobra.Command {
	opts := serveOptions{}

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Validate and prepare the MCP stdio server",
		Long:  "Validate PocketBase MCP stdio server configuration and environment fallback before startup.",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := promptPocketBaseCredentials(cmd, opts.URL, opts.Email, opts.Password)
			if err != nil {
				return err
			}
			opts.URL = creds.URL
			opts.Email = creds.Email
			opts.Password = creds.Password

			resolved, err := config.ResolveServeConfig(config.ServeConfigInput{
				URL:       opts.URL,
				Email:     opts.Email,
				Password:  opts.Password,
				TimeoutMS: opts.TimeoutMS,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "mcp config valid: url=%s email=%s timeout_ms=%d password=%s\n",
				resolved.URL,
				resolved.Email,
				resolved.TimeoutMS,
				config.RedactedPassword(resolved.Password),
			)
			fmt.Fprintln(cmd.OutOrStdout(), "next: implement PocketBase auth, MCP tool registry, and stdio transport wiring")
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.URL, "url", "", "PocketBase URL")
	cmd.Flags().StringVar(&opts.Email, "email", "", "PocketBase user email")
	cmd.Flags().StringVar(&opts.Email, "user", "", "Alias for --email")
	cmd.Flags().StringVar(&opts.Password, "password", "", "PocketBase password")
	cmd.Flags().IntVar(&opts.TimeoutMS, "timeout-ms", 0, "Request timeout in milliseconds")

	return cmd
}
