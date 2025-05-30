/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	gothci_cli "github.com/felipegenef/gothic-cli/CLI"
	"github.com/spf13/cobra"
)

// hotReloadCmd represents the hotReload command
var hotReloadCmd = &cobra.Command{
	Use:   "hot-reload",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: newHotReloadComand(gothci_cli.NewCli()),
}

func newHotReloadComand(cli gothci_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		comand := gothci_cli.NewHotReloadCommandCli(&cli)
		comand.HotReload()
		return nil
	}
}

func init() {
	rootCmd.AddCommand(hotReloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hotReloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hotReloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
