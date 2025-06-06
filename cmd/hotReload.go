/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	templGenerate "github.com/a-h/templ/cmd/templ/generatecmd"
	gothic_cli "github.com/felipegenef/gothic-cli/pkg/cli"
	gothic_helpers "github.com/felipegenef/gothic-cli/pkg/helpers"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var hotReloadCmd = &cobra.Command{
	Use:   "hot-reload",
	Short: "Run your Gothic app locally in hot-reload mode.",
	Long: `This command uses Templ and Tailwind to enable real-time reloading for local development.

It allows you to develop and debug your Gothic app more efficiently, with changes instantly reflected in the browser as you save your files.`,
	RunE: newHotReloadCommand(gothic_cli.NewCli()),
}

func init() {
	rootCmd.AddCommand(hotReloadCmd)
}

type HotReloadCommand struct {
	cli               *gothic_cli.GothicCli
	tailwindFile      string
	mainBinaryName    string
	runCmd            *exec.Cmd
	runCancel         context.CancelFunc
	mutex             sync.Mutex
	excludedDirs      []string
	watchedExtensions []string
	excludeRegex      regexp.Regexp
}

func newHotReloadCommandCli(cli *gothic_cli.GothicCli) HotReloadCommand {
	var tailwindBinary string = "./tailwindcss"
	var mainBinary string = "tmp/main"
	if runtime.GOOS == "windows" {
		tailwindBinary = "./tailwindcss.exe"
		mainBinary = "tmp/main.exe"
	}
	return HotReloadCommand{
		cli:               cli,
		tailwindFile:      tailwindBinary,
		mainBinaryName:    mainBinary,
		excludedDirs:      []string{"assets", "tmp", "vendor", "public", "routes"},
		watchedExtensions: []string{".go", ".tpl", ".tmpl", ".templ", ".html"},
		excludeRegex:      *regexp.MustCompile(`.*_templ\.go$`),
	}
}

func newHotReloadCommand(cli gothic_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		command := newHotReloadCommandCli(&cli)

		return command.HotReload()
	}
}

func (command *HotReloadCommand) HotReload() error {
	go command.watchTailwindChanges()
	// Wait for tailwind process to render css for the first time
	time.Sleep(4 * time.Second)
	go command.watchForChanges()
	go command.watchTemplChanges()

	banner := `
 ██████╗  ██████╗ ████████╗██╗  ██╗██╗ ██████╗     █████╗ ██████╗ ██████╗ 
██╔════╝ ██╔═══██╗╚══██╔══╝██║  ██║██║██╔════╝    ██╔══██╗██╔══██╗██╔══██╗
██║  ███╗██║   ██║   ██║   ███████║██║██║         ███████║██████╔╝██████╔╝
██║   ██║██║   ██║   ██║   ██╔══██║██║██║         ██╔══██║██╔═══╝ ██╔═══╝ 
╚██████╔╝╚██████╔╝   ██║   ██║  ██║██║╚██████╗    ██║  ██║██║     ██║     
 ╚═════╝  ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚═╝ ╚═════╝    ╚═╝  ╚═╝╚═╝     ╚═╝     

🚀 Gothic App is up and running!
🌐 Listening on: http://127.0.0.1:7331
🔥  Mode: HOT RELOAD ENABLED
`
	fmt.Println(banner)

	select {}

}

func (command *HotReloadCommand) isExcludedDir(path string) bool {
	for _, d := range command.excludedDirs {
		if strings.Contains(path, string(os.PathSeparator)+d+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (command *HotReloadCommand) watchForChanges() {
	command.rebuild()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("error creating watcher: %v", err)
	}
	defer watcher.Close()
	err = filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && command.isExcludedDir(path) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking through directories: %v", err)
		command.rebuild()
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if command.shouldHandle(event.Name, event.Op) {
				command.rebuild()
			}
			// Dynamically watch new directories
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() && !command.isExcludedDir(event.Name) {
					err := watcher.Add(event.Name)
					if err == nil {
						log.Printf("New directory added to watcher: %s", event.Name)
					} else {
						log.Printf("Failed to add new directory to watcher: %s, error: %v", event.Name, err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			command.rebuild()
			if !ok {
				fmt.Printf("error reloading app: %v", err)
			}
			log.Println("Watcher error:", err)
		}
	}
}

func (command *HotReloadCommand) shouldHandle(path string, op fsnotify.Op) bool {
	if command.isExcludedDir(path) {
		return false
	}

	filename := filepath.Base(path)
	if command.excludeRegex.MatchString(filename) {
		// Ignore templ-generated files unless they are deleted
		if op&(fsnotify.Remove) == 0 {
			return false
		}
	}

	ext := filepath.Ext(path)
	for _, e := range command.watchedExtensions {
		if e == ext {
			return true
		}
	}
	return false
}

func (command *HotReloadCommand) watchTailwindChanges() {
	log.Println("Starting Tailwind in watch mode...")

	tailWindCmd := exec.Command(command.tailwindFile, "--watch=always", "-i", "src/css/app.css", "-o", "public/styles.css", "--minify")
	tailWindCmd.Stdout = os.Stdout
	tailWindCmd.Stderr = os.Stderr

	// Start the process asynchronously (non-blocking, like Node's spawn)
	if err := tailWindCmd.Start(); err != nil {
		fmt.Printf("Failed to start Tailwind watch process: %v", err)
	}

	log.Printf("Tailwind is watching with PID %d", tailWindCmd.Process.Pid)

	// Optionally wait for the process to exit and log its exit
	go func() {
		err := tailWindCmd.Wait()
		if err != nil {
			fmt.Printf("Tailwind process exited with error: %v", err)
		} else {
			log.Println("Tailwind process exited normally.")
		}
	}()
}

func (command *HotReloadCommand) watchTemplChanges() {
	logger := gothic_helpers.NewLogger("error", false, os.Stdout)

	templGenerate.Run(context.Background(), logger, templGenerate.Arguments{
		Watch:       true,
		Proxy:       "http://localhost:8080",
		OpenBrowser: true,
	})
}

func (command *HotReloadCommand) rebuild() {
	command.mutex.Lock()
	defer command.mutex.Unlock()

	log.Println("Build routes...")
	if err := command.cli.FileBasedRouter.Render(command.cli.GetConfig().GoModName); err != nil {
		fmt.Printf("error building routes: %v", err)
		return
	}

	log.Println("Build app...")
	buildCmd := exec.Command("go", "build", "-o", command.mainBinaryName, "main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("error building app: %v", err)
		return
	}

	if command.runCancel != nil {
		log.Println("Stopping previous go run process...")
		command.runCancel()
		command.runCancel = nil
	}
	log.Println("Running app...")
	ctx, cancel := context.WithCancel(context.Background())
	command.runCancel = cancel

	runCmd := exec.CommandContext(ctx, command.mainBinaryName)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	command.runCmd = runCmd
	go func() {
		if err := runCmd.Run(); err != nil {
			if ctx.Err() == nil {
				fmt.Printf("error running app: %v", err)
			}
		}
	}()

}
