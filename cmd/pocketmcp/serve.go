package pocketmcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mreyeswilson/pocketmcp/internal/config"
	"github.com/mreyeswilson/pocketmcp/internal/mcpserver"
	"github.com/mreyeswilson/pocketmcp/internal/pocketbase"
	"github.com/spf13/cobra"
	"time"
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
		Short: "Run the PocketBase MCP stdio server",
		Long:  "Start an MCP stdio server authenticated as PocketBase superuser for collections, records, and settings administration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if canPrompt() {
				creds, err := promptPocketBaseCredentials(cmd, opts.URL, opts.Email, opts.Password)
				if err != nil {
					return err
				}
				opts.URL = creds.URL
				opts.Email = creds.Email
				opts.Password = creds.Password
			}

			resolved, err := config.ResolveServeConfig(config.ServeConfigInput{
				URL:       opts.URL,
				Email:     opts.Email,
				Password:  opts.Password,
				TimeoutMS: opts.TimeoutMS,
			})
			if err != nil {
				return err
			}

			pb := pocketbase.NewClient(
				resolved.URL,
				resolved.Email,
				resolved.Password,
				time.Duration(resolved.TimeoutMS)*time.Millisecond,
			)
			server := mcpserver.New(pb, version)
			fmt.Fprintf(cmd.ErrOrStderr(), "PocketBase MCP server running: url=%s email=%s timeout_ms=%d version=%s\n",
				resolved.URL,
				resolved.Email,
				resolved.TimeoutMS,
				version,
			)
			return server.Run(context.Background(), &mcp.StdioTransport{})
		},
	}

	cmd.Flags().StringVar(&opts.URL, "url", "", "PocketBase URL")
	cmd.Flags().StringVar(&opts.Email, "email", "", "PocketBase user email")
	cmd.Flags().StringVar(&opts.Email, "user", "", "Alias for --email")
	cmd.Flags().StringVar(&opts.Password, "password", "", "PocketBase password")
	cmd.Flags().IntVar(&opts.TimeoutMS, "timeout-ms", 0, "Request timeout in milliseconds")

	return cmd
}
