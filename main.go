package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"

	cli_utils "github.com/felipegenef/gothic-cli/utils"
	"github.com/teris-io/shortid"
)

//go:embed .gothicCli
var gothicCliFolder embed.FS

//go:embed server
var serverFolder embed.FS

//go:embed public
var publicFolder embed.FS

//go:embed src
var srcFolder embed.FS

//go:embed makefile
var makeFile embed.FS

//go:embed tailwind.config.js
var tailwindConfig embed.FS

//go:embed README.md
var readme embed.FS

//go:embed gothic-config.json
var goticConfig embed.FS

//go:embed tailwindcss-linux
var tailwindCSSLinux embed.FS

//go:embed tailwindcss-windows.exe
var tailwindCSSWindows embed.FS

//go:embed tailwindcss-mac
var tailwindCSSMac embed.FS

var airToml string = `root = "."
tmp_dir = "tmp"

[build]
  bin = "{{.MainBinaryFileName}}"
  cmd = "./{{.TailWindFileName}} -i src/css/app.css -o public/styles.css --minify && templ generate && go build -o {{.MainBinaryFileName}} main.go"
    
  delay = 2
  exclude_dir = ["assets", "tmp", "vendor","public"]
  exclude_file = []
  exclude_regex = [".*_templ.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "templ", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[log]
  time = false

[misc]
  clean_on_exit = false`

var envs string = `HTTP_LISTEN_ADDR: ":8080"
LOCAL_SERVE: "true"`

var gitIgnore string = `.env
bin
*_templ.go*
node_modules
.aws-sam
tmp
optimize/*
public/styles.css
template.yaml
samconfig.toml
Dockerfile`

var rootDirs = []string{
	"public",
	".gothicCli",
	"src",
	"optimize",
}
var cliDirs = []string{
	".gothicCli/HotReload",
	".gothicCli/imgOptimization",
	".gothicCli/imgOptimization/setup",
	".gothicCli/buildSamTemplate",
	".gothicCli/buildSamTemplate/templates",
	".gothicCli/buildSamTemplate/cleanup",
	".gothicCli/CdnAddOrRemoveAssets",
	".gothicCli/sam",
}
var srcDirs = []string{
	"src/api",
	"src/components",
	"src/css",
	"src/layouts",
	"src/pages",
	"src/utils",
}
var publicDirs = []string{
	"public/imageExample",
}

var cliFiles = map[string]embed.FS{
	".gothicCli/HotReload/main.go":       gothicCliFolder,
	".gothicCli/imgOptimization/main.go": gothicCliFolder,
	".gothicCli/shared.go":               gothicCliFolder,
	".gothicCli/buildSamTemplate/templates/template-custom-domain-with-arn.yaml": gothicCliFolder,
	".gothicCli/buildSamTemplate/templates/template-custom-domain.yaml":          gothicCliFolder,
	".gothicCli/buildSamTemplate/templates/template-default.yaml":                gothicCliFolder,
	".gothicCli/buildSamTemplate/templates/samconfig-template.toml":              gothicCliFolder,
	".gothicCli/buildSamTemplate/main.go":                                        gothicCliFolder,
	".gothicCli/buildSamTemplate/templates/Dockerfile-template":                  gothicCliFolder,
	".gothicCli/buildSamTemplate/cleanup/main.go":                                gothicCliFolder,
	".gothicCli/CdnAddOrRemoveAssets/main.go":                                    gothicCliFolder,
	".gothicCli/sam/main.go":                                                     gothicCliFolder,
	".gothicCli/imgOptimization/setup/main.go":                                   gothicCliFolder,
}

var publicFolderFiles = map[string]embed.FS{
	"public/imageExample/blurred.jpeg":  publicFolder,
	"public/imageExample/original.jpeg": publicFolder,
	"public/favicon.ico":                publicFolder,
}

var rootFiles = map[string]embed.FS{
	"makefile":           makeFile,
	"tailwind.config.js": tailwindConfig,
	"README.md":          readme,
	"gothic-config.json": goticConfig,
}

var apiFiles = map[string]embed.FS{
	"src/api/helloWorld.go": srcFolder,
}

var componentFiles = map[string]embed.FS{
	"src/components/helloWorld.templ":     srcFolder,
	"src/components/optimizedImage.templ": srcFolder,
	"src/components/lazyLoad.templ":       srcFolder,
}

var cssFiles = map[string]embed.FS{
	"src/css/app.css": srcFolder,
}

var layoutFiles = map[string]embed.FS{
	"src/layouts/layout.templ": srcFolder,
}

var pageFiles = map[string]embed.FS{
	"src/pages/index.templ":      srcFolder,
	"src/pages/revalidate.templ": srcFolder,
}

var utilFiles = map[string]embed.FS{
	"src/utils/handler.go": srcFolder,
}

var globalRequiredLibs = []string{"github.com/a-h/templ/cmd/templ", "github.com/air-verse/air"}

var templateFiles = map[string]string{
	"src/pages/index.templ":                   "src/pages/index.templ",
	"src/pages/revalidate.templ":              "src/pages/revalidate.templ",
	".gothicCli/CdnAddOrRemoveAssets/main.go": ".gothicCli/CdnAddOrRemoveAssets/main.go",
	".gothicCli/sam/main.go":                  ".gothicCli/sam/main.go",
	".gothicCli/imgOptimization/main.go":      ".gothicCli/imgOptimization/main.go",
	".gothicCli/buildSamTemplate/main.go":     ".gothicCli/buildSamTemplate/main.go",
	"gothic-config.json":                      "gothic-config.json",
}

func main() {
	currentRuntime := runtime.GOOS
	initCmd := flag.Bool("init", false, "Initialize project files and directories")
	buildCmd := flag.String("build", "", "Build project (options: page, static-page, isr-page, api-route, isr-api-route, component, isr-component, lazy-load-component)")
	helpCmd := flag.Bool("help", false, "Display help information")
	flag.Parse()

	if *helpCmd {
		displayHelp()
		return
	}

	if *initCmd {
		projectName := promptForProjectName()
		goModName := promptForGoModName()
		if projectName == "" {
			fmt.Println("Project name cannot be empty.")
			return
		}

		if err := initializeProject(projectName, goModName, currentRuntime); err != nil {
			fmt.Printf("Error initializing the project: %v\n", err)
		} else {
			initializeModule(goModName)
			templ := exec.Command("make", "templ")
			templ.Run()
			gitinit := exec.Command("git", "init")
			gitinit.Run()
			fmt.Println("Project initialized successfully!")
		}
	} else if *buildCmd != "" {
		name := promptForBuildName(*buildCmd)
		if name != "" {
			handleBuild(*buildCmd, name)
		}
	} else {
		fmt.Println("Use --init to initialize the project or --build to build a boilerplate example.")
	}

}

func initializeModule(goModuleName string) {
	initCmd := exec.Command("go", "mod", "init", goModuleName)
	initCmd.Stdin = os.Stdin
	initCmd.Stderr = os.Stderr
	initCmd.Run()
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdin = os.Stdin
	tidyCmd.Stderr = os.Stderr
	tidyCmd.Run()
	installRequiredLibs()
	checkForUpdates()
}

func installRequiredLibs() {
	fmt.Println("Installing dependencies...")
	for _, lib := range globalRequiredLibs {
		cmd := exec.Command("go", "install", lib+"@latest")
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

func checkForUpdates() {
	fmt.Println("Checking for updates on dependencies...")
	for _, lib := range globalRequiredLibs {
		cmd := exec.Command("go", "get", "-u", lib)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

// Function to create directories and files
func initializeProject(projectName string, goModName string, currentRuntime string) error {
	initCmdTemplateInfo := cli_utils.InitCMDTemplateInfo{ProjectName: projectName, GoModName: goModName, MainServerPackageName: "package main", MainServerFunctionName: "main()"}
	upperId, err := shortid.Generate()
	if err != nil {
		fmt.Println("Error generating short ID:", err)
		return fmt.Errorf("Error generating app id")
	}
	// Replace all special characters with -
	re := regexp.MustCompile(`[^\w\s]|_`)
	lowerId := strings.ToLower(upperId)
	id := re.ReplaceAllString(lowerId, "-")

	for _, dir := range rootDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	for _, dir := range publicDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	for _, dir := range cliDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	for _, dir := range srcDirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	// Create Tailwind with special permissions for execution based on OS
	switch currentRuntime {
	case "linux":
		data, _ := fs.ReadFile(tailwindCSSLinux, "tailwindcss-linux")
		initCmdTemplateInfo.TailWindFileName = "tailwindcss"
		initCmdTemplateInfo.MainBinaryFileName = "./temp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {

			return fmt.Errorf("error creating file %s: %w", "tailwindcss", err)
		}
	case "darwin":
		data, _ := fs.ReadFile(tailwindCSSMac, "tailwindcss-mac")
		initCmdTemplateInfo.TailWindFileName = "tailwindcss"
		initCmdTemplateInfo.MainBinaryFileName = "./temp/main"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss", data, 0755); err != nil {

			return fmt.Errorf("error creating file %s: %w", "tailwindcss", err)
		}
	case "windows":
		data, _ := fs.ReadFile(tailwindCSSWindows, "tailwindcss-windows.exe")
		initCmdTemplateInfo.TailWindFileName = "tailwindcss.exe"
		initCmdTemplateInfo.MainBinaryFileName = "./temp/main.exe"
		// Write the file with executable permissions (0755)
		if err := os.WriteFile("tailwindcss.exe", data, 0755); err != nil {

			return fmt.Errorf("error creating file %s: %w", "tailwindcss", err)
		}
	default:
		fmt.Println("Unknown OS.")
	}

	// Create dot files (embed api wont let dots on files)
	go func(initCmdTemplateInfo cli_utils.InitCMDTemplateInfo) {
		os.WriteFile(".air.toml", []byte(airToml), 0644)
		cli_utils.ReplaceOnFile(".air.toml", ".air.toml", initCmdTemplateInfo)
	}(initCmdTemplateInfo)

	go func(appId string) {

		os.WriteFile(".gothicCli/app-id.txt", []byte(appId), 0644)
	}(id)

	go func() {
		os.WriteFile(".env", []byte(envs), 0644)
	}()

	go func() {
		os.WriteFile(".gitignore", []byte(gitIgnore), 0644)
	}()

	// Create and replace package on serverfile
	mainServerData, _ := fs.ReadFile(serverFolder, "server/server.go")
	if err := os.WriteFile("main.go", mainServerData, 0644); err != nil {

		return fmt.Errorf("error creating file %s: %w", "main.go", err)
	}
	cli_utils.ReplaceOnFile("main.go", "main.go", initCmdTemplateInfo)

	if err := createFiles(cliFiles); err != nil {
		return err
	}
	if err := createFiles(publicFolderFiles); err != nil {
		return err
	}
	if err := createFiles(rootFiles); err != nil {
		return err
	}
	// Replace tailwind bin name on scripts
	cli_utils.ReplaceOnFile("makefile", "makefile", initCmdTemplateInfo)
	if err := createFiles(apiFiles); err != nil {
		return err
	}
	if err := createFiles(componentFiles); err != nil {
		return err
	}
	if err := createFiles(cssFiles); err != nil {
		return err
	}
	if err := createFiles(layoutFiles); err != nil {
		return err
	}
	if err := createFiles(pageFiles); err != nil {
		return err
	}
	if err := createFiles(utilFiles); err != nil {
		return err
	}

	for templateFilePath, outputPath := range templateFiles {
		go func(templateFilePath, outputPath string, initCmdTemplateInfo cli_utils.InitCMDTemplateInfo) {
			err := cli_utils.ReplaceOnFile(templateFilePath, outputPath, initCmdTemplateInfo)
			if err != nil {
				fmt.Errorf("error generating file %s: %w", outputPath, err)
			}

		}(templateFilePath, outputPath, initCmdTemplateInfo)

	}

	return nil
}

// Function to create files in parallel
func createFiles(files map[string]embed.FS) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(files)) // Channel to collect errors

	for filename, fileContent := range files {
		wg.Add(1) // Increment the WaitGroup counter
		go func(filename string, fileContent embed.FS) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			data, err := fs.ReadFile(fileContent, filename)
			if err != nil {
				errChan <- fmt.Errorf("error reading file %s: %w", filename, err)
				return
			}

			if err := os.WriteFile(filename, data, 0644); err != nil {
				errChan <- fmt.Errorf("error creating file %s: %w", filename, err)
				return
			}
		}(filename, fileContent) // Pass parameters to the goroutine
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan) // Close the error channel

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func handleBuild(buildType, name string) error {
	switch buildType {
	case "page":
		if err := buildAndReplace(name, srcFolder, "src/pages/index.templ", "Index", "src/pages/"+name+".templ"); err != nil {
			return err
		}

		originalRouteExample := `Please add this to your api routes to use the component:

			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				handler.Render(r, w, pages.Index())
			})`
		templ := exec.Command("make", "templ")
		templ.Run()
		fmt.Println(strings.ReplaceAll(originalRouteExample, "Index", name))
	case "static-page":

		if err := buildAndReplace(name, srcFolder, "src/pages/index.templ", "Index", "src/pages/"+name+".templ"); err != nil {
			return err
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

	case "isr-page":

		if err := buildAndReplace(name, srcFolder, "src/pages/revalidate.templ", "Revalidate", "src/pages/"+name+".templ"); err != nil {
			return err
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
	case "api-route":

		if err := buildAndReplace(name, srcFolder, "src/api/helloWorld.go", "HelloWorld", "src/api/"+name+".go"); err != nil {
			return err
		}
		templ := exec.Command("make", "templ")
		templ.Run()
		originalRouteExample := `Please add this to your api routes to use the component:

				router.Get("/", api.HelloWorld)`

		fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))

	case "isr-api-route":
		if err := buildAndReplace(name, srcFolder, "src/api/helloWorld.go", "HelloWorld", "src/api/"+name+".go"); err != nil {
			return err
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
	case "component":
		if err := buildAndReplace(name, srcFolder, "src/components/helloWorld.templ", "HelloWorld", "src/components/"+name+".templ"); err != nil {
			return err
		}
		templ := exec.Command("make", "templ")
		templ.Run()
		originalRouteExample := `Please add this to your api routes to use the component:

				router.Get("/", func(w http.ResponseWriter, r *http.Request) {
					handler.Render(r, w, components.HelloWorld())
				})`

		fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))
	case "isr-component":
		if err := buildAndReplace(name, srcFolder, "src/components/helloWorld.templ", "HelloWorld", "src/components/"+name+".templ"); err != nil {
			return err
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
	case "lazy-load-component":
		if err := buildAndReplace(name, srcFolder, "src/components/lazyLoad.templ", "LazyLoad", "src/components/"+name+".templ"); err != nil {
			return err
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
	default:
		fmt.Println("Unknown build type. Use one of: page, static-page, isr-page, api-route, isr-api-route, component, isr-component, lazy-load-component.")
	}

	return nil
}
func buildAndReplace(name string, fileTemplate embed.FS, fileTemplatePath string, stringToReplace string, outputFilePath string) error {
	data, err := fs.ReadFile(fileTemplate, fileTemplatePath)
	if err != nil {
		return err
	}
	replacedData := []byte(strings.ReplaceAll(string(data), stringToReplace, name))

	if err := os.WriteFile(outputFilePath, replacedData, 0644); err != nil {
		return err
	}

	return nil
}
func displayHelp() {
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

func promptForProjectName() string {
	var name string
	fmt.Print("Enter your unique stack name in kebab case (e.g., your-unique-stack-name): ")
	fmt.Scanln(&name)

	// Validate kebab case
	if matched, _ := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, name); !matched {
		fmt.Println("Invalid name format. Please use kebab case (lowercase letters and numbers only, with dashes).")
		return ""
	}
	return name
}

func promptForGoModName() string {
	var name string
	fmt.Print("Enter your go module name: ")
	fmt.Scanln(&name)
	return name
}

func promptForBuildName(buildType string) string {
	var name string
	fmt.Printf("Enter the name for the %s (in camel case, e.g., MyPageExample,MyApiRoute,MyComponent etc...): ", buildType)
	fmt.Scanln(&name)

	// Validate camel case
	if matched, _ := regexp.MatchString(`^[A-Z][a-zA-Z0-9]*$`, name); !matched {
		fmt.Println("Invalid name format. Please use camel case (start with uppercase letter, followed by letters and digits).")
		return ""
	}
	return name
}
