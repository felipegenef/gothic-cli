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
	Short: "Compiles all Templ files into go files.",
	Long:  `Internal command intented to be called before deploy and between hot reloads to build golang files from templ files.`,
	RunE:  newBuildCommand(gothic_cli.NewCli()),
}

func newBuildCommand(cli gothic_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		return cli.Templ.Render()
	}
}

func init() {
	rootCmd.AddCommand(buildCmd)

}
