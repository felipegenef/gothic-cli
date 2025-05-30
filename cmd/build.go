/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"

	templGenerate "github.com/a-h/templ/cmd/templ/generatecmd"
	gothci_cli "github.com/felipegenef/gothic-cli/CLI"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := gothci_cli.NewLogger("error", false, os.Stdout)

		err := templGenerate.Run(context.Background(), logger, templGenerate.Arguments{})
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

}
