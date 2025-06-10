/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	gothci_cli "github.com/felipegenef/gothicframework/pkg/cli"
	cli_data "github.com/felipegenef/gothicframework/pkg/data"
	helpers "github.com/felipegenef/gothicframework/pkg/helpers"
	"github.com/spf13/cobra"
	"github.com/teris-io/shortid"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project structure and configuration files for a Gothic app.",
	Long: `Sets up the initial folder structure and essential files required to start building a Gothic app.

This includes:
  - A precompiled Tailwind binary
  - A gothic-config.json file
  - A basic example app to help you get started
  - A link to the official documentation for further guidance`,
	RunE: newInitCommand(gothci_cli.NewCli()),
}

func init() {
	rootCmd.AddCommand(initCmd)
}

type InitCommand struct {
	gothicCliData cli_data.GothicCliData
	cli           *gothci_cli.GothicCli
}

func NewInitCommandCli(cli *gothci_cli.GothicCli, gothicCliData cli_data.GothicCliData) InitCommand {
	return InitCommand{
		cli:           cli,
		gothicCliData: gothicCliData,
	}
}

func newInitCommand(cli gothci_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		command := NewInitCommandCli(&cli, cli_data.DefaultCLIData)

		return command.CreateNewGothicApp(cli_data.DefaultCLIData)
	}
}

func (command *InitCommand) CreateNewGothicApp(data cli_data.GothicCliData) error {

	projectName, err := command.promptForProjectName()
	if err != nil {
		return err
	}
	data.ProjectName = projectName
	gomodName, err := command.promptForGoModName()
	if err != nil {
		return err
	}
	data.GoModName = gomodName
	command.gothicCliData = data

	command.initializeProject()
	command.cli.InitializeModule(command.gothicCliData.GoModName)
	if err := command.cli.Templ.Render(); err != nil {
		return err
	}

	if err := command.cli.FileBasedRouter.Render(gomodName); err != nil {
		return err
	}

	gitinit := exec.Command("git", "init")
	gitinit.Run()
	fmt.Println("Project initialized successfully!")
	return nil
}

// Function to create directories and files
func (command *InitCommand) initializeProject() error {

	command.cli.Templates.InitCmdTemplateInfo = helpers.InitCmdTemplateInfo{
		ProjectName:            command.gothicCliData.ProjectName,
		GoModName:              command.gothicCliData.GoModName,
		MainServerPackageName:  "package main",
		MainServerFunctionName: "main()",
	}

	if err := command.createInitialDirs(); err != nil {
		return err
	}
	// Create Tailwind with special permissions for execution based on OS
	if err := command.createTailwindBinary(); err != nil {
		return err
	}
	// Create dot files (embed api wont let dots on files)
	if err := command.createHiddenFiles(); err != nil {
		return err
	}

	// Create initial file structure
	if err := command.createInitialFileStructure(); err != nil {
		return err
	}
	// create all custom template files
	if err := command.createTemplateBasedFiles(); err != nil {
		return err
	}
	return nil
}

func (command *InitCommand) createInitialDirs() error {
	for _, dir := range command.gothicCliData.InitialDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("error generating initial Directories: %v", err)
		}
	}
	return nil
}

func (command *InitCommand) createTailwindBinary() error {
	switch command.cli.Runtime {
	case "linux":
		data, _ := fs.ReadFile(command.gothicCliData.Tailwind.Linux, "tailwindcss-linux")
		command.cli.Templates.InitCmdTemplateInfo.TailWindFileName = "tailwindcss"
		command.cli.Templates.InitCmdTemplateInfo.MainBinaryFileName = "./tmp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {
			return fmt.Errorf("error creating file %s: %w", command.cli.Templates.InitCmdTemplateInfo.TailWindFileName, err)
		}
	case "darwin":
		data, _ := fs.ReadFile(command.gothicCliData.Tailwind.Mac, "tailwindcss-mac")
		command.cli.Templates.InitCmdTemplateInfo.TailWindFileName = "tailwindcss"
		command.cli.Templates.InitCmdTemplateInfo.MainBinaryFileName = "./tmp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {
			return fmt.Errorf("error creating file %s: %w", command.cli.Templates.InitCmdTemplateInfo.TailWindFileName, err)
		}
	case "windows":
		data, _ := fs.ReadFile(command.gothicCliData.Tailwind.Windows, "tailwindcss-windows.exe")
		command.cli.Templates.InitCmdTemplateInfo.TailWindFileName = "tailwindcss.exe"
		command.cli.Templates.InitCmdTemplateInfo.MainBinaryFileName = "./tmp/main.exe"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss.exe", data, 0755); err != nil {
			return fmt.Errorf("error creating file %s: %w", command.cli.Templates.InitCmdTemplateInfo.TailWindFileName, err)
		}
	default:
		return fmt.Errorf("error: unknown OS")
	}
	return nil
}

func (command *InitCommand) createHiddenFiles() error {

	upperId, err := shortid.Generate()
	if err != nil {
		return fmt.Errorf("error generating app ID: %v", err)
	}
	// Replace all special characters with -
	re := regexp.MustCompile(`[^\w\s]|_`)
	lowerId := strings.ToLower(upperId)
	id := re.ReplaceAllString(lowerId, "-")

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		os.WriteFile(".gothicCli/app-id.txt", []byte(id), 0644)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".env", []byte(command.gothicCliData.Env), 0644)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".gitignore", []byte(command.gothicCliData.GitIgnore), 0644)
		wg.Done()
	}()
	wg.Wait()
	return nil

}

func (command *InitCommand) createInitialFileStructure() error {
	mainServerData, _ := fs.ReadFile(command.gothicCliData.ServerFolder, "server/server.go")
	if err := os.WriteFile("main.go", mainServerData, 0644); err != nil {
		return fmt.Errorf("error creating file %s: %w", "main.go", err)
	}
	command.cli.Templates.UpdateFromTemplate("main.go", "main.go", command.cli.Templates.InitCmdTemplateInfo)

	var wg sync.WaitGroup

	for filename, fileContent := range command.gothicCliData.InitialFiles {
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := command.cli.Templates.CreateFromTemplate(fileContent, filename, filename, command.cli.Templates.InitCmdTemplateInfo); err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", "main.go", err))
			}
		}()
	}

	for filename, fileContent := range command.gothicCliData.TemplateFiles {
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := command.cli.Templates.CopyFromFs(fileContent, filename, filename); err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", "main.go", err))
			}
		}()
	}

	for filename, fileContent := range command.gothicCliData.PublicFolderAssets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := fs.ReadFile(fileContent, filename)
			if err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", filename, err))
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", filename, err))
			}
		}()
	}

	wg.Wait()
	return nil
}

func (command *InitCommand) createTemplateBasedFiles() error {
	var wg sync.WaitGroup

	wg.Add(len(command.gothicCliData.CustomTemplateBasedPages) + len(command.gothicCliData.CustomTemplateBasedComponents) + len(command.gothicCliData.CustomTemplateBasedRoutes))

	// Pages
	for templateFilePath, pageName := range command.gothicCliData.CustomTemplateBasedPages {
		go func() {
			if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, templateFilePath, templateFilePath, helpers.RouteTemplateInfo{PageName: pageName, GoModName: command.gothicCliData.GoModName}); err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", templateFilePath, err))
			}
			wg.Done()
		}()

	}

	// Components
	for templateFilePath, componentName := range command.gothicCliData.CustomTemplateBasedComponents {
		go func() {
			if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, templateFilePath, templateFilePath, helpers.RouteTemplateInfo{ComponentName: componentName, GoModName: command.gothicCliData.GoModName}); err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", templateFilePath, err))
			}
			wg.Done()
		}()

	}

	// API Routes
	for templateFilePath, routeName := range command.gothicCliData.CustomTemplateBasedRoutes {
		go func() {
			if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, templateFilePath, templateFilePath, helpers.RouteTemplateInfo{RouteName: routeName, GoModName: command.gothicCliData.GoModName}); err != nil {
				panic(fmt.Sprintf("error creating file %s: %v", templateFilePath, err))
			}
			wg.Done()
		}()

	}
	wg.Wait()
	return nil

}

func (command *InitCommand) promptForProjectName() (string, error) {
	var name string
	fmt.Print("Enter your unique stack name in kebab case (e.g., your-unique-stack-name): ")
	fmt.Scanln(&name)

	// Validate kebab case
	if matched, _ := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, name); !matched {
		return "", fmt.Errorf("invalid name format. Please use kebab case (lowercase letters and numbers only, with dashes)")
	}
	if name == "" {
		return "", fmt.Errorf("project name cannot be empty")
	}
	return name, nil
}

func (command *InitCommand) promptForGoModName() (string, error) {
	var name string
	fmt.Print("Enter your go module name: ")
	fmt.Scanln(&name)
	if name == "" {
		return "", fmt.Errorf("go module name cannot be empty")
	}
	return name, nil
}
