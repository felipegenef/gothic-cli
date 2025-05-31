/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	gothic_cli "github.com/felipegenef/gothic-cli/pkg/cli"
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
	RunE: newBuildCommand(gothic_cli.NewCli()),
}

func newBuildCommand(cli gothic_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		err := cli.Templ.Render()
		return err
	}
}

func init() {
	rootCmd.AddCommand(buildCmd)

}
