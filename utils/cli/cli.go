package cli

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"

	gothicCliShared "github.com/felipegenef/gothic-cli/.gothicCli"
	templates_cli "github.com/felipegenef/gothic-cli/utils/templates"
	"github.com/teris-io/shortid"
)

type GothicCli struct {
	config             *gothicCliShared.Config
	GlobalRequiredLibs []string
	runtime            string
	createAppData      GothicCliData
	Templates          templates_cli.Templates
}

type CliCommands struct {
	Init  *bool
	Build *string
	Help  *bool
}

type TailWindCSS struct {
	Mac     embed.FS
	Windows embed.FS
	Linux   embed.FS
	Config  embed.FS
}

type GothicCliData struct {
	InitialFiles                  map[string]embed.FS
	PublicFolderAssets            map[string]embed.FS
	InitialDirs                   []string
	CustomTemplateBasedPages      map[string]string
	CustomTemplateBasedComponents map[string]string
	CustomTemplateBasedRoutes     map[string]string
	GitIgnore                     string
	EnvSample                     string
	Env                           string
	AirToml                       string
	Tailwind                      TailWindCSS
	GoticConfig                   embed.FS
	Readme                        embed.FS
	MakeFile                      embed.FS
	SrcFolder                     embed.FS
	PublicFolder                  embed.FS
	ServerFolder                  embed.FS
	GothicCliFolder               embed.FS
	ProjectName                   string
	GoModName                     string
}

func NewCli() GothicCli {

	return GothicCli{
		GlobalRequiredLibs: []string{"github.com/a-h/templ/cmd/templ", "github.com/air-verse/air"},
		runtime:            runtime.GOOS,
		Templates:          templates_cli.NewCLITemplate(),
	}
}

func (cli *GothicCli) WaitForCommands() CliCommands {
	initCmd := flag.Bool("init", false, "Initialize project files and directories")
	buildCmd := flag.String("build", "", "Build project (options: page, static-page, isr-page, api-route, isr-api-route, component, isr-component, lazy-load-component)")
	helpCmd := flag.Bool("help", false, "Display help information")
	flag.Parse()

	return CliCommands{Init: initCmd, Build: buildCmd, Help: helpCmd}
}

func (cli *GothicCli) PromptHelpInstructions() {
	fmt.Println("Usage:")
	fmt.Println("  --init                     Initialize project files and directories.")
	fmt.Println("  --build <type>            Build project with specified type.")
	fmt.Println("                           Options for <type>:")
	fmt.Println("                             page           Build a regular page.")
	fmt.Println("                             static-page    Build a static page.")
	fmt.Println("                             isr-page       Build an ISR page.")
	fmt.Println("                             api-route      Build an API route.")
	fmt.Println("                             isr-api-route  Build an ISR API route.")
	fmt.Println("  --help                     Display this help information.")
}

func (cli *GothicCli) BuildCommand(buildCmd *string, data GothicCliData) {
	name := cli.promptBuildCommandName(*buildCmd)
	cli.createAppData = data
	if name != "" {
		cli.handleBuild(*buildCmd, name)
	}
}

func (cli *GothicCli) GetConfig() gothicCliShared.Config {
	if cli.config != nil {
		return *cli.config
	}
	var config gothicCliShared.Config
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

func (cli *GothicCli) CreateNewGothicApp(data GothicCliData) {

	data.ProjectName = cli.promptForProjectName()
	data.GoModName = cli.promptForGoModName()
	cli.createAppData = data

	cli.initializeProject()
	cli.initializeModule(cli.createAppData.GoModName)
	templ := exec.Command("make", "templ")
	templ.Run()
	gitinit := exec.Command("git", "init")
	gitinit.Run()
	fmt.Println("Project initialized successfully!")
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

// Function to create directories and files
func (cli *GothicCli) initializeProject() {

	cli.Templates.InitCMDTemplateInfo = templates_cli.InitCMDTemplateInfo{
		ProjectName:            cli.createAppData.ProjectName,
		GoModName:              cli.createAppData.GoModName,
		MainServerPackageName:  "package main",
		MainServerFunctionName: "main()",
	}

	cli.createInitialDirs()
	// Create Tailwind with special permissions for execution based on OS
	cli.createTailwindBinary()
	// Create dot files (embed api wont let dots on files)
	cli.createHiddenFiles()

	// Create initial file structure
	cli.createInitialFileStructure()
	// create all custom template files
	cli.createTemplateBasedFiles()
}

func (cli *GothicCli) createTemplateBasedFiles() {
	var wg sync.WaitGroup

	wg.Add(len(cli.createAppData.CustomTemplateBasedPages) + len(cli.createAppData.CustomTemplateBasedComponents) + len(cli.createAppData.CustomTemplateBasedRoutes))

	// Pages
	for templateFilePath, pageName := range cli.createAppData.CustomTemplateBasedPages {
		go func() {
			if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, templateFilePath, templateFilePath, templates_cli.BuildCMDTemplateInfo{PageName: pageName, GoModName: cli.createAppData.GoModName}); err != nil {
				log.Fatalf("error creating file %s: %w", templateFilePath, err)
			}
			wg.Done()
		}()

	}

	// Components
	for templateFilePath, componentName := range cli.createAppData.CustomTemplateBasedComponents {
		go func() {
			if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, templateFilePath, templateFilePath, templates_cli.BuildCMDTemplateInfo{ComponentName: componentName, GoModName: cli.createAppData.GoModName}); err != nil {
				log.Fatalf("error creating file %s: %w", templateFilePath, err)
			}
			wg.Done()
		}()

	}

	// API Routes
	for templateFilePath, routeName := range cli.createAppData.CustomTemplateBasedRoutes {
		go func() {
			if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, templateFilePath, templateFilePath, templates_cli.BuildCMDTemplateInfo{RouteName: routeName, GoModName: cli.createAppData.GoModName}); err != nil {
				log.Fatalf("error creating file %s: %w", templateFilePath, err)
			}
			wg.Done()
		}()

	}
	wg.Wait()
}

func (cli *GothicCli) createInitialFileStructure() {
	mainServerData, _ := fs.ReadFile(cli.createAppData.ServerFolder, "server/server.go")
	if err := os.WriteFile("main.go", mainServerData, 0644); err != nil {
		log.Fatalf("error creating file %s: %w", "main.go", err)
	}
	cli.Templates.UpdateFromTemplate("main.go", "main.go", cli.Templates.InitCMDTemplateInfo)

	var wg sync.WaitGroup

	for filename, fileContent := range cli.createAppData.InitialFiles {
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := cli.Templates.CreateFromTemplate(fileContent, filename, filename, cli.Templates.InitCMDTemplateInfo); err != nil {
				log.Fatalf("error creating file %s: %w", filename, err)
			}
		}()
	}

	for filename, fileContent := range cli.createAppData.PublicFolderAssets {
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

func (cli *GothicCli) createHiddenFiles() {

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
		os.WriteFile(".air.toml", []byte(cli.createAppData.AirToml), 0644)
		cli.Templates.UpdateFromTemplate(".air.toml", ".air.toml", cli.Templates.InitCMDTemplateInfo)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".gothicCli/app-id.txt", []byte(id), 0644)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".env", []byte(cli.createAppData.Env), 0644)
		os.WriteFile(".env.sample", []byte(cli.createAppData.EnvSample), 0644)
		wg.Done()
	}()

	go func() {
		os.WriteFile(".gitignore", []byte(cli.createAppData.GitIgnore), 0644)
		wg.Done()
	}()
	wg.Wait()

}

func (cli *GothicCli) createInitialDirs() {
	for _, dir := range cli.createAppData.InitialDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("Error generating initial Directories:", err)
		}
	}
}

func (cli *GothicCli) createTailwindBinary() {
	switch cli.runtime {
	case "linux":
		data, _ := fs.ReadFile(cli.createAppData.Tailwind.Linux, "tailwindcss-linux")
		cli.Templates.InitCMDTemplateInfo.TailWindFileName = "tailwindcss"
		cli.Templates.InitCMDTemplateInfo.MainBinaryFileName = "./temp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {
			log.Fatalf("error creating file %s: %w", cli.Templates.InitCMDTemplateInfo.TailWindFileName, err)
		}
	case "darwin":
		data, _ := fs.ReadFile(cli.createAppData.Tailwind.Mac, "tailwindcss-mac")
		cli.Templates.InitCMDTemplateInfo.TailWindFileName = "tailwindcss"
		cli.Templates.InitCMDTemplateInfo.MainBinaryFileName = "./temp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {
			log.Fatalf("error creating file %s: %w", cli.Templates.InitCMDTemplateInfo.TailWindFileName, err)
		}
	case "windows":
		data, _ := fs.ReadFile(cli.createAppData.Tailwind.Windows, "tailwindcss-windows.exe")
		cli.Templates.InitCMDTemplateInfo.TailWindFileName = "tailwindcss.exe"
		cli.Templates.InitCMDTemplateInfo.MainBinaryFileName = "./temp/main.exe"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss.exe", data, 0755); err != nil {
			log.Fatalf("error creating file %s: %w", cli.Templates.InitCMDTemplateInfo.TailWindFileName, err)
		}
	default:
		fmt.Println("Unknown OS.")
	}

}

func (cli *GothicCli) handleBuild(buildType, name string) error {
	cli.GetConfig()
	cli.Templates.BuildCMDTemplateInfo = templates_cli.BuildCMDTemplateInfo{
		PageName:      name,
		RouteName:     name,
		ComponentName: name,
		GoModName:     cli.config.GoModName,
	}

	switch buildType {
	case "page":
		cli.buildPage(name)
	case "static-page":
		cli.buildStaticPage(name)
	case "isr-page":
		cli.buildISRPage(name)
	case "api-route":
		cli.buildApiRoute(name)
	case "isr-api-route":
		cli.buildIsrApiRoute(name)
	case "component":
		cli.buildComponent(name)
	case "isr-component":
		cli.buildIsrComponent(name)
	case "lazy-load-component":
		cli.buildLazyLoadComponent(name)
	default:
		fmt.Println("Unknown build type. Use one of: page, static-page, isr-page, api-route, isr-api-route, component, isr-component, lazy-load-component.")
	}

	return nil
}

func (cli *GothicCli) promptForProjectName() string {
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

func (cli *GothicCli) promptForGoModName() string {
	var name string
	fmt.Print("Enter your go module name: ")
	fmt.Scanln(&name)
	if name == "" {
		log.Fatalln("go module name cannot be empty.")
	}
	return name
}

func (cli *GothicCli) promptBuildCommandName(buildCmd string) string {
	var name string
	fmt.Printf("Enter the name for the %s (in camel case, e.g., MyPageExample,MyApiRoute,MyComponent etc...): ", buildCmd)
	fmt.Scanln(&name)

	// Validate camel case
	if matched, _ := regexp.MatchString(`^[A-Z][a-zA-Z0-9]*$`, name); !matched {
		fmt.Println("Invalid name format. Please use camel case (start with uppercase letter, followed by letters and digits).")
	}
	return name
}

func (cli *GothicCli) buildPage(name string) {
	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/pages/index.templ", "src/pages/"+name+".templ", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating page: %w", err)
	}

	originalRouteExample := `Please add this to your api routes to use the component:

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handler.Render(r, w, pages.Index())
		})`
	templ := exec.Command("make", "templ")
	templ.Run()
	fmt.Println(strings.ReplaceAll(originalRouteExample, "Index", name))
}

func (cli *GothicCli) buildStaticPage(name string) {
	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/pages/index.templ", "src/pages/"+name+".templ", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating static page: %w", err)
	}

	originalRouteExample := `Please add this to your api routes to use the component:
	
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// Max cache age for CloudFront is 31536000 = 1 year
			w.Header().Set("Cache-Control", "max-age=31536000")
			handler.Render(r, w, pages.Index())
		})`
	templ := exec.Command("make", "templ")
	templ.Run()
	fmt.Println(strings.ReplaceAll(originalRouteExample, "Index", name))
}

func (cli *GothicCli) buildISRPage(name string) {
	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/pages/revalidate.templ", "src/pages/"+name+".templ", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating isr-page: %w", err)
	}

	originalRouteExample := `Please add this to your api routes to use the component:

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// Revalidate page every 10 seconds. You can revalidate up to 31536000 (1 year)
			w.Header().Set("Cache-Control", "max-age=10, stale-while-revalidate=10, stale-if-error=10")
			handler.Render(r, w, pages.Index())
		})`
	templ := exec.Command("make", "templ")
	templ.Run()
	fmt.Println(strings.ReplaceAll(originalRouteExample, "Revalidate", name))
}

func (cli *GothicCli) buildApiRoute(name string) {

	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/api/helloWorld.go", "src/api/"+name+".go", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating Api Route: %w", err)
	}
	templ := exec.Command("make", "templ")
	templ.Run()
	originalRouteExample := `Please add this to your api routes to use the component:

			router.Get("/", api.HelloWorld)`

	fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))
}

func (cli *GothicCli) buildIsrApiRoute(name string) {

	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/api/helloWorld.go", "src/api/"+name+".go", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating Api Route: %w", err)
	}

	originalRouteExample := `Please add this to your api routes to use the component:
	
			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				// Revalidate page every 10 seconds. You can revalidate up to 31536000 (1 year)
				w.Header().Set("Cache-Control", "max-age=10, stale-while-revalidate=10, stale-if-error=10")
				api.HelloWorld(w,r)
			})`
	templ := exec.Command("make", "templ")
	templ.Run()
	fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))
}

func (cli *GothicCli) buildComponent(name string) {
	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/components/helloWorld.templ", "src/components/"+name+".templ", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating Api Route: %w", err)
	}
	templ := exec.Command("make", "templ")
	templ.Run()
	originalRouteExample := `Please add this to your api routes to use the component:

			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				handler.Render(r, w, components.HelloWorld())
			})`

	fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))
}

func (cli *GothicCli) buildIsrComponent(name string) {
	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/components/helloWorld.templ", "src/components/"+name+".templ", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating Api Route: %w", err)
	}
	templ := exec.Command("make", "templ")
	templ.Run()
	originalRouteExample := `Please add this to your api routes to use the component:

			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				// Revalidate page every 10 seconds. You can revalidate up to 31536000 (1 year)
				w.Header().Set("Cache-Control", "max-age=10, stale-while-revalidate=10, stale-if-error=10")
				handler.Render(r, w, components.HelloWorld())
			})`

	fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))
}

func (cli *GothicCli) buildLazyLoadComponent(name string) {
	if err := cli.Templates.CreateFromTemplate(cli.createAppData.SrcFolder, "src/components/lazyLoad.templ", "src/components/"+name+".templ", cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating Api Route: %w", err)
	}
	templ := exec.Command("make", "templ")
	templ.Run()
	originalRouteExample := `Please add this to your api routes to use the component:

			router.Get("/yourLazyLoadedComponent", func(w http.ResponseWriter, r *http.Request) {
				handler.Render(r, w, components.LazyLoad(false))
			})
	
	
			Also add this to your page to lazy load the component
			
			@components.LazyLoad(true)
`

	fmt.Println(strings.ReplaceAll(originalRouteExample, "LazyLoad", name))
}
