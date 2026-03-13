package pocketmcp

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:           "pocketmcp",
	Short:         "PocketBase MCP CLI",
	Long:          "PocketBase MCP CLI implemented in Go with Cobra.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newMCPCmd())
	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newUninstallCmd())
	rootCmd.AddCommand(newVersionCmd())
}
