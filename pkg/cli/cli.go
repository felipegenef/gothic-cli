package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	helpers "github.com/felipegenef/gothicframework/pkg/helpers"
	routes "github.com/felipegenef/gothicframework/pkg/helpers/routes"
)

type GothicCli struct {
	config  *Config
	appID   *string
	Runtime string

	Templates       helpers.TemplateHelper
	Tailwind        helpers.TailwindHelper
	Templ           helpers.TemplHelper
	Logger          *slog.Logger
	AwsSam          helpers.AwsSamHelper
	AWS             helpers.AwsHelper
	FileBasedRouter routes.FileBasedRouteHelper
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

		Runtime:         runtime.GOOS,
		Templates:       helpers.NewTemplateHelper(),
		Tailwind:        helpers.NewTailwindHelper(),
		Templ:           helpers.NewTemplHelper(),
		AwsSam:          helpers.NewAwsSamHelper(),
		AWS:             helpers.NewAwsHelper(),
		Logger:          helpers.NewLogger("error", false, os.Stdout),
		FileBasedRouter: routes.NewFileBasedRouteHelper(),
	}

	return cli
}

func (cli *GothicCli) GetAppId() (string, error) {
	if cli.appID != nil {
		return *cli.appID, nil
	}
	content, err := os.ReadFile(".gothicCli/app-id.txt")
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Convert the content to string
	appID := string(content)
	cli.appID = &appID
	return appID, nil
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
