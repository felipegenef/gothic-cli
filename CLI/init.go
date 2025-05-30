package cli

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	cli_data "github.com/felipegenef/gothic-cli/data"
	"github.com/teris-io/shortid"
)

type InitCommand struct {
	gothicCliData cli_data.GothicCliData
	cli           *GothicCli
}

func NewInitCommandCli(cli *GothicCli, gothicCliData cli_data.GothicCliData) InitCommand {
	return InitCommand{
		cli:           cli,
		gothicCliData: gothicCliData,
	}
}

func (command *InitCommand) CreateNewGothicApp(data cli_data.GothicCliData) {

	data.ProjectName = command.promptForProjectName()
	data.GoModName = command.promptForGoModName()
	command.gothicCliData = data

	command.initializeProject()
	command.cli.initializeModule(command.gothicCliData.GoModName)
	templ := exec.Command("make", "templ")
	templ.Run()
	gitinit := exec.Command("git", "init")
	gitinit.Run()
	fmt.Println("Project initialized successfully!")
}

// Function to create directories and files
func (command *InitCommand) initializeProject() {

	command.cli.Templates.InitCMDTemplateInfo = InitCMDTemplateInfo{
		ProjectName:            command.gothicCliData.ProjectName,
		GoModName:              command.gothicCliData.GoModName,
		MainServerPackageName:  "package main",
		MainServerFunctionName: "main()",
	}

	command.createInitialDirs()
	// Create Tailwind with special permissions for execution based on OS
	command.createTailwindBinary()
	// Create dot files (embed api wont let dots on files)
	command.createHiddenFiles()

	// Create initial file structure
	command.createInitialFileStructure()
	// create all custom template files
	command.createTemplateBasedFiles()
}

func (command *InitCommand) createInitialDirs() {
	for _, dir := range command.gothicCliData.InitialDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("Error generating initial Directories:", err)
		}
	}
}

func (command *InitCommand) createTailwindBinary() {
	switch command.cli.runtime {
	case "linux":
		data, _ := fs.ReadFile(command.gothicCliData.Tailwind.Linux, "tailwindcss-linux")
		command.cli.Templates.InitCMDTemplateInfo.TailWindFileName = "tailwindcss"
		command.cli.Templates.InitCMDTemplateInfo.MainBinaryFileName = "./tmp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {
			log.Fatalf("error creating file %s: %w", command.cli.Templates.InitCMDTemplateInfo.TailWindFileName, err)
		}
	case "darwin":
		data, _ := fs.ReadFile(command.gothicCliData.Tailwind.Mac, "tailwindcss-mac")
		command.cli.Templates.InitCMDTemplateInfo.TailWindFileName = "tailwindcss"
		command.cli.Templates.InitCMDTemplateInfo.MainBinaryFileName = "./tmp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {
			log.Fatalf("error creating file %s: %w", command.cli.Templates.InitCMDTemplateInfo.TailWindFileName, err)
		}
	case "windows":
		data, _ := fs.ReadFile(command.gothicCliData.Tailwind.Windows, "tailwindcss-windows.exe")
		command.cli.Templates.InitCMDTemplateInfo.TailWindFileName = "tailwindcss.exe"
		command.cli.Templates.InitCMDTemplateInfo.MainBinaryFileName = "./tmp/main.exe"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss.exe", data, 0755); err != nil {
			log.Fatalf("error creating file %s: %w", command.cli.Templates.InitCMDTemplateInfo.TailWindFileName, err)
		}
	default:
		fmt.Println("Unknown OS.")
	}

}

func (command *InitCommand) createHiddenFiles() {

	upperId, err := shortid.Generate()
	if err != nil {
		log.Fatalf("Error generating short ID:", err)
	}
	// Replace all special characters with -
	re := regexp.MustCompile(`[^\w\s]|_`)
	lowerId := strings.ToLower(upperId)
	id := re.ReplaceAllString(lowerId, "-")

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		os.WriteFile(".air.toml", []byte(command.gothicCliData.AirToml), 0644)
		command.cli.Templates.UpdateFromTemplate(".air.toml", ".air.toml", command.cli.Templates.InitCMDTemplateInfo)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".gothicCli/app-id.txt", []byte(id), 0644)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".env", []byte(command.gothicCliData.Env), 0644)
		os.WriteFile(".env.sample", []byte(command.gothicCliData.EnvSample), 0644)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".gitignore", []byte(command.gothicCliData.GitIgnore), 0644)
		wg.Done()
	}()
	wg.Wait()

}

func (command *InitCommand) createInitialFileStructure() {
	mainServerData, _ := fs.ReadFile(command.gothicCliData.ServerFolder, "server/server.go")
	if err := os.WriteFile("main.go", mainServerData, 0644); err != nil {
		log.Fatalf("error creating file %s: %w", "main.go", err)
	}
	command.cli.Templates.UpdateFromTemplate("main.go", "main.go", command.cli.Templates.InitCMDTemplateInfo)

	var wg sync.WaitGroup

	for filename, fileContent := range command.gothicCliData.InitialFiles {
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := command.cli.Templates.CreateFromTemplate(fileContent, filename, filename, command.cli.Templates.InitCMDTemplateInfo); err != nil {
				log.Fatalf("error creating file %s: %w", filename, err)
			}
		}()
	}

	for filename, fileContent := range command.gothicCliData.PublicFolderAssets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := fs.ReadFile(fileContent, filename)
			if err != nil {
				log.Fatalf("error creating file %s: %w", filename, err)
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				log.Fatalf("error creating file %s: %w", filename, err)
			}
		}()
	}

	wg.Wait()
}

func (command *InitCommand) createTemplateBasedFiles() {
	var wg sync.WaitGroup

	wg.Add(len(command.gothicCliData.CustomTemplateBasedPages) + len(command.gothicCliData.CustomTemplateBasedComponents) + len(command.gothicCliData.CustomTemplateBasedRoutes))

	// Pages
	for templateFilePath, pageName := range command.gothicCliData.CustomTemplateBasedPages {
		go func() {
			if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, templateFilePath, templateFilePath, BuildCMDTemplateInfo{PageName: pageName, GoModName: command.gothicCliData.GoModName}); err != nil {
				log.Fatalf("error creating file %s: %w", templateFilePath, err)
			}
			wg.Done()
		}()

	}

	// Components
	for templateFilePath, componentName := range command.gothicCliData.CustomTemplateBasedComponents {
		go func() {
			if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, templateFilePath, templateFilePath, BuildCMDTemplateInfo{ComponentName: componentName, GoModName: command.gothicCliData.GoModName}); err != nil {
				log.Fatalf("error creating file %s: %w", templateFilePath, err)
			}
			wg.Done()
		}()

	}

	// API Routes
	for templateFilePath, routeName := range command.gothicCliData.CustomTemplateBasedRoutes {
		go func() {
			if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, templateFilePath, templateFilePath, BuildCMDTemplateInfo{RouteName: routeName, GoModName: command.gothicCliData.GoModName}); err != nil {
				log.Fatalf("error creating file %s: %w", templateFilePath, err)
			}
			wg.Done()
		}()

	}
	wg.Wait()
}

func (command *InitCommand) promptForProjectName() string {
	var name string
	fmt.Print("Enter your unique stack name in kebab case (e.g., your-unique-stack-name): ")
	fmt.Scanln(&name)

	// Validate kebab case
	if matched, _ := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, name); !matched {
		log.Fatalln("Invalid name format. Please use kebab case (lowercase letters and numbers only, with dashes).")
	}
	if name == "" {
		log.Fatalln("Project name cannot be empty.")
	}
	return name
}

func (command *InitCommand) promptForGoModName() string {
	var name string
	fmt.Print("Enter your go module name: ")
	fmt.Scanln(&name)
	if name == "" {
		log.Fatalln("go module name cannot be empty.")
	}
	return name
}
