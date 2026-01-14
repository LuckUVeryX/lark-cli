package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/config"
	"github.com/yjwong/lark-cli/internal/output"
)

var rootCmd = &cobra.Command{
	Use:   "lark",
	Short: "Lark CLI for Claude Code",
	Long: `A CLI tool to interact with Lark APIs.
Designed for use by Claude Code with JSON output.

All commands output JSON by default.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := config.Init(); err != nil {
		output.Fatal("CONFIG_ERROR", err)
	}

	if err := rootCmd.Execute(); err != nil {
		output.Fatal("COMMAND_ERROR", err)
	}
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(calCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(contactCmd)
	rootCmd.AddCommand(docCmd)
	rootCmd.AddCommand(msgCmd)
}
