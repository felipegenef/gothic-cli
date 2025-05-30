/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	gothci_cli "github.com/felipegenef/gothic-cli/CLI"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: newDeployComand(gothci_cli.NewCli()),
}

func newDeployComand(cli gothci_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		comand := gothci_cli.NewDeployCommandCli(&cli)
		stageFlag, err := cmd.Flags().GetString("stage")
		if err != nil {
			return err
		}
		action, err := cmd.Flags().GetString("action")
		if err != nil {
			return err
		}
		if !isValidAction(action) {
			return fmt.Errorf("error: invalid action \"%s\". Allowed values: %v", action, allowedActions)

		}
		comand.Deploy(stageFlag, action)
		return nil
	}
}

var allowedActions = []string{"delete", "deploy"}

func isValidAction(c string) bool {
	for _, a := range allowedActions {
		if a == c {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("stage", "s", "dev", "Define AWS stage to deploy or delete")
	deployCmd.Flags().StringP("action", "a", "deploy", "Action to be taken, to deploy or delete the api")
}
