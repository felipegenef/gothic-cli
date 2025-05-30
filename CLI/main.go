package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

type GothicCli struct {
	config             *Config
	GlobalRequiredLibs []string
	runtime            string

	Templates              Templates
	InitCommand            InitCommand
	DeployCommand          DeployCommand
	HotReloadCommand       HotReloadCommand
	ImgOptimizationCommand ImgOptimizationCommand
}

type CliCommands struct {
	Init            *bool
	Build           *string
	Deploy          *bool
	Help            *bool
	ImgOptimization *bool
	HotReload       *bool
	DeployAction    *string
	DeployStage     *string
}

func NewCli() GothicCli {
	cli := GothicCli{
		GlobalRequiredLibs: []string{},
		runtime:            runtime.GOOS,
	}
	// Referene cli pointer for children
	cli.InitCommand.cli = &cli
	cli.DeployCommand.cli = &cli
	cli.HotReloadCommand.cli = &cli
	cli.ImgOptimizationCommand.cli = &cli

	return cli
}

func (cli *GothicCli) GetConfig() Config {
	if cli.config != nil {
		return *cli.config
	}
	var config Config
	file, err := os.Open("gothic-config.json")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
		panic(err)
	}
	defer file.Close()

	// Decode the JSON from the file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
		panic(err)
	}
	cli.config = &config
	return config
}

func (cli *GothicCli) initializeModule(goModuleName string) {
	initCmd := exec.Command("go", "mod", "init", goModuleName)
	initCmd.Stdin = os.Stdin
	initCmd.Stderr = os.Stderr
	initCmd.Run()
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdin = os.Stdin
	tidyCmd.Stderr = os.Stderr
	tidyCmd.Run()
	cli.InstallRequiredLibs()
	cli.CheckForUpdates()
}

func (cli *GothicCli) InstallRequiredLibs() {
	fmt.Println("Installing dependencies...")
	for _, lib := range cli.GlobalRequiredLibs {
		cmd := exec.Command("go", "install", lib+"@latest")
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

func (cli *GothicCli) CheckForUpdates() {
	fmt.Println("Checking for updates on dependencies...")
	for _, lib := range cli.GlobalRequiredLibs {
		cmd := exec.Command("go", "get", "-u", lib)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
