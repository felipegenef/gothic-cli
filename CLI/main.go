package cli

import (
	"encoding/json"
	"flag"
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
	BuildCommand           BuildCommand
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
		GlobalRequiredLibs: []string{"github.com/a-h/templ/cmd/templ", "github.com/air-verse/air"},
		runtime:            runtime.GOOS,
	}
	// Referene cli pointer for children
	cli.BuildCommand.cli = &cli
	cli.InitCommand.cli = &cli
	cli.DeployCommand.cli = &cli
	cli.HotReloadCommand.cli = &cli
	cli.ImgOptimizationCommand.cli = &cli

	return cli
}

func (cli *GothicCli) WaitForCommands() CliCommands {
	initCmd := flag.Bool("init", false, "Initialize project files and directories")
	buildCmd := flag.String("build", "", "Build project (options: page, static-page, isr-page, api-route, isr-api-route, component, isr-component, lazy-load-component)")
	helpCmd := flag.Bool("help", false, "Display help information")
	imgOptimizationCmd := flag.Bool("optimize-images", false, "Display help information")
	hotReloadCmd := flag.Bool("hot-reload", false, "Display help information")
	deployCmd := flag.Bool("deploy", false, "Specify the deployment stage (default, dev, staging, prod)")
	deployActionFlag := flag.String("action", "deploy", "Ether deploy or delete the app")
	deployStageFlag := flag.String("stage", "default", "Specify the deployment stage (default, dev, staging, prod)")

	flag.Parse()

	return CliCommands{Init: initCmd, Build: buildCmd, Help: helpCmd, ImgOptimization: imgOptimizationCmd, HotReload: hotReloadCmd, Deploy: deployCmd, DeployAction: deployActionFlag, DeployStage: deployStageFlag}
}

func (cli *GothicCli) PromptHelpInstructions() {
	fmt.Println("Usage:")
	fmt.Println("  --init                  Initialize project files and directories.")
	fmt.Println("  --build <type>         Build the project with the specified type.")
	fmt.Println("                         Options for <type>:")
	fmt.Println("                           page             Build a regular page.")
	fmt.Println("                           static-page      Build a static page.")
	fmt.Println("                           isr-page         Build an ISR page.")
	fmt.Println("                           api-route        Build an API route.")
	fmt.Println("                           isr-api-route    Build an ISR API route.")
	fmt.Println("  --hot-reload           Start hot reloading for the current app.")
	fmt.Println("  --optimize-images      Generate images in different resolutions from originals in the optimize folder.")
	fmt.Println("  --deploy <stage>       Deploy the app to the specified stage (default, dev, staging, prod).")
	fmt.Println("  --help                 Display this help information.")
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
