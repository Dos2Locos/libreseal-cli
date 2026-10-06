package cmd

import (
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "🥷  Integrate LibreSeal with AI Agents",
	Long:  "Configure how AI coding agents interact with your LibreSeal secrets.",
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
