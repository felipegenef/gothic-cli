package cli

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	helpers "github.com/felipegenef/gothic-cli/pkg/helpers"
)

type GothicCli struct {
	config  *Config
	Runtime string

	Templates helpers.TemplateHelper
	Tailwind  helpers.TailwindHelper
	Templ     helpers.TemplHelper
	Air       helpers.AirHelper
	Logger    *slog.Logger
	AwsSam    helpers.AwsSamHelper
	AWS       helpers.AwsHelper
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

		Runtime:   runtime.GOOS,
		Templates: helpers.NewTemplateHelper(),
		Tailwind:  helpers.NewTailwindHelper(),
		Templ:     helpers.NewTemplHelper(),
		Air:       helpers.NewAirHelper(),
		AwsSam:    helpers.NewAwsSamHelper(),
		AWS:       helpers.NewAwsHelper(),
		Logger:    helpers.NewLogger("error", false, os.Stdout),
	}

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

func (cli *GothicCli) InitializeModule(goModuleName string) {
	initCmd := exec.Command("go", "mod", "init", goModuleName)
	initCmd.Stdin = os.Stdin
	initCmd.Stderr = os.Stderr
	initCmd.Run()
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdin = os.Stdin
	tidyCmd.Stderr = os.Stderr
	tidyCmd.Run()
}
