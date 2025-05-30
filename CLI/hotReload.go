package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	templGenerate "github.com/a-h/templ/cmd/templ/generatecmd"
	runner "github.com/air-verse/air/runner"
)

type HotReloadCommand struct {
	cli *GothicCli
}

func NewHotReloadCommandCli(cli *GothicCli) HotReloadCommand {

	return HotReloadCommand{
		cli: cli,
	}
}

func (command *HotReloadCommand) HotReload() {
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

	go func() {
		logger := NewLogger("error", false, os.Stdout)

		templGenerate.Run(context.Background(), logger, templGenerate.Arguments{
			Watch: true,
			Proxy: "http://localhost:8080",
		})
	}()
	time.Sleep(3 * time.Second)

	// time.Sleep(4 * time.Second)
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
}
