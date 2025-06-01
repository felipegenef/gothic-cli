/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	templGenerate "github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/air-verse/air/runner"
	gothic_cli "github.com/felipegenef/gothic-cli/pkg/cli"
	gothic_helpers "github.com/felipegenef/gothic-cli/pkg/helpers"
	"github.com/spf13/cobra"
)

var hotReloadCmd = &cobra.Command{
	Use:   "hot-reload",
	Short: "Run your Gothic app locally in hot-reload mode.",
	Long: `This command uses Templ, Air, and Tailwind to enable real-time reloading for local development.

It allows you to develop and debug your Gothic app more efficiently, with changes instantly reflected in the browser as you save your files.`,
	RunE: newHotReloadComand(gothic_cli.NewCli()),
}

func newHotReloadComand(cli gothic_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		comand := newHotReloadCommandCli(&cli)

		return comand.HotReload()
	}
}

func init() {
	rootCmd.AddCommand(hotReloadCmd)
}

type HotReloadCommand struct {
	cli *gothic_cli.GothicCli
}

func newHotReloadCommandCli(cli *gothic_cli.GothicCli) HotReloadCommand {

	return HotReloadCommand{
		cli: cli,
	}
}

func (command *HotReloadCommand) HotReload() error {
	go func() {

		cfg, err := runner.InitConfig("")
		if err != nil {
			log.Fatal(err)
			return
		}
		r, err := runner.NewEngineWithConfig(cfg, false)
		if err != nil {
			log.Fatal(err)
			return
		}
		r.Run()
	}()
	time.Sleep(3 * time.Second)
	go func() {
		logger := gothic_helpers.NewLogger("error", false, os.Stdout)

		templGenerate.Run(context.Background(), logger, templGenerate.Arguments{
			Watch: true,
			Proxy: "http://localhost:8080",
		})
	}()
	time.Sleep(2 * time.Second)

	banner := `
 ██████╗  ██████╗ ████████╗██╗  ██╗██╗ ██████╗     █████╗ ██████╗ ██████╗ 
██╔════╝ ██╔═══██╗╚══██╔══╝██║  ██║██║██╔════╝    ██╔══██╗██╔══██╗██╔══██╗
██║  ███╗██║   ██║   ██║   ███████║██║██║         ███████║██████╔╝██████╔╝
██║   ██║██║   ██║   ██║   ██╔══██║██║██║         ██╔══██║██╔═══╝ ██╔═══╝ 
╚██████╔╝╚██████╔╝   ██║   ██║  ██║██║╚██████╗    ██║  ██║██║     ██║     
 ╚═════╝  ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚═╝ ╚═════╝    ╚═╝  ╚═╝╚═╝     ╚═╝     

🚀 Gothic App is up and running!
🌐 Listening on: http://127.0.0.1:7331
♻️  Mode: HOT RELOAD ENABLED
`
	fmt.Println(banner)

	select {}
	return nil
}
