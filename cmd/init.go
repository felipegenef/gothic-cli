/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	gothci_cli "github.com/felipegenef/gothic-cli/CLI"
	cli_data "github.com/felipegenef/gothic-cli/data"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project files and directories.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: newInitComand(gothci_cli.NewCli()),
}

func newInitComand(cli gothci_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		comand := gothci_cli.NewInitCommandCli(&cli, cli_data.DefaultCLIData)
		comand.CreateNewGothicApp(cli_data.DefaultCLIData)
		return nil
	}
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
