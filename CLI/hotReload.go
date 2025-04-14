package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type HotReloadCommand struct {
	cli *GothicCli
}

var (
	templCmd *exec.Cmd
	airCmd   *exec.Cmd
	mu       sync.Mutex
)

func NewHotReloadCommandCli() HotReloadCommand {
	return HotReloadCommand{}
}

func (command *HotReloadCommand) HotReload() {
	// Start the commands and wait for their completion
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		command.runCommand("templ", []string{"generate", "--watch", "--proxy=http://localhost:8080"})
	}()
	// TODO remove air and replace for custom go script
	go func() {
		defer wg.Done()
		command.runCommand("air", []string{})
	}()
	select {}
}

func (command *HotReloadCommand) runCommand(name string, args []string) {
	cmd := exec.Command(name, args...)
	mu.Lock()
	if name == "templ" {
		templCmd = cmd
	} else if name == "air" {
		airCmd = cmd
	}
	mu.Unlock()

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		command.handleError(fmt.Sprintf("Error creating stdout pipe for command %s", name), err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		command.handleError(fmt.Sprintf("Error creating stderr pipe for command %s", name), err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		command.handleError(fmt.Sprintf("Error starting command %s", name), err)
		return
	}

	// Print stdout
	go io.Copy(os.Stdout, stdout)

	// Print stderr
	go io.Copy(os.Stderr, stderr)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		command.handleError(fmt.Sprintf("Command %s finished with error", name), err)
	}
}

func (command *HotReloadCommand) handleError(message string, err error) {
	log.Println(message, err)
	command.gracefulShutdown()
}

func (command *HotReloadCommand) gracefulShutdown() {
	mu.Lock()
	defer mu.Unlock()

	if templCmd != nil {
		templCmd.Process.Kill()
	}
	if airCmd != nil {
		airCmd.Process.Kill()
	}

	os.Exit(1)
}
