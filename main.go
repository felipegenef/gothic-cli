package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

//go:embed Dockerfile
var dockerFile embed.FS

//go:embed go.mod
var goMod embed.FS

//go:embed go.sum
var goSum embed.FS

//go:embed server/server.go
var mainServerFile embed.FS

//go:embed makefile
var makeFile embed.FS

//go:embed samconfig.toml
var samConfigTom embed.FS

//go:embed tailwind.config.js
var tailwindConfig embed.FS

//go:embed tailwindcss
var tailwindCSS embed.FS

//go:embed template.yaml
var templateYaml embed.FS

//go:embed README.md
var readme embed.FS

//go:embed CLI/HotReload/main.go
var hotReloadScript embed.FS

//go:embed CLI/imgOptimization/main.go
var imgOptimizationScript embed.FS

//go:embed public/imageExample/blurred.jpeg
var imgOptimizationBlurredImg embed.FS

//go:embed public/imageExample/original.jpeg
var imgOptimizationOriginalImg embed.FS

//go:embed public/favicon.ico
var favicon embed.FS

//go:embed src/api/helloWorld.go
var apiExample embed.FS

//go:embed src/components/helloWorld.templ
var helloWorldExample embed.FS

//go:embed src/components/lazyLoad.templ
var lazyLoadExample embed.FS

//go:embed src/components/optimizedImage.templ
var optimizeImageExample embed.FS

//go:embed src/css/app.css
var appCSS embed.FS

//go:embed src/layouts/layout.templ
var layout embed.FS

//go:embed src/pages/index.templ
var indexPage embed.FS

//go:embed src/pages/revalidate.templ
var revalidatePage embed.FS

//go:embed src/utils/handler.go
var utils embed.FS

var airToml string = `root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "./tailwindcss -i src/css/app.css -o public/styles.css --minify && templ generate && go build -o ./tmp/main main.go"
    
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
public/styles.css`

var rootDirs = []string{"public", "CLI", "src", "optimize"}
var cliDirs = []string{"CLI/HotReload", "CLI/imgOptimization"}
var srcDirs = []string{"src/api", "src/components", "src/css", "src/layouts", "src/pages", "src/utils"}
var publicDirs = []string{"public/imageExample"}

var cliFiles = map[string]embed.FS{
	"CLI/HotReload/main.go":       hotReloadScript,
	"CLI/imgOptimization/main.go": imgOptimizationScript,
}

var publicFolderFiles = map[string]embed.FS{
	"public/imageExample/blurred.jpeg":  imgOptimizationBlurredImg,
	"public/imageExample/original.jpeg": imgOptimizationOriginalImg,
	"public/favicon.ico":                favicon,
}

var srcFiles = map[string]embed.FS{
	"Dockerfile":         dockerFile,
	"go.mod":             goMod,
	"go.sum":             goSum,
	"makefile":           makeFile,
	"samconfig.toml":     samConfigTom,
	"tailwind.config.js": tailwindConfig,
	"template.yaml":      templateYaml,
	"README.md":          readme,
}

var apiFiles = map[string]embed.FS{
	"src/api/helloWorld.go": apiExample,
}

var componentFiles = map[string]embed.FS{
	"src/components/helloWorld.templ":     helloWorldExample,
	"src/components/optimizedImage.templ": optimizeImageExample,
	"src/components/lazyLoad.templ":       lazyLoadExample,
}

var cssFiles = map[string]embed.FS{
	"src/css/app.css": appCSS,
}

var layoutFiles = map[string]embed.FS{
	"src/layouts/layout.templ": layout,
}

var pageFiles = map[string]embed.FS{
	"src/pages/index.templ":      indexPage,
	"src/pages/revalidate.templ": revalidatePage,
}

var utilFiles = map[string]embed.FS{
	"src/utils/handler.go": utils,
}

func main() {
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
		if projectName == "" {
			fmt.Println("Project name cannot be empty.")
			return
		}

		if err := initializeProject(projectName); err != nil {
			fmt.Printf("Error initializing the project: %v\n", err)
		} else {
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

// Function to create directories and files
func initializeProject(projectName string) error {
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

	// Create dot files (embed api wont let dots on files)
	go func() {
		os.WriteFile(".air.toml", []byte(airToml), 0644)
	}()

	go func() {
		os.WriteFile(".env", []byte(envs), 0644)
	}()

	go func() {
		os.WriteFile(".env.sample", []byte(envs), 0644)
	}()

	go func() {
		os.WriteFile(".gitignore", []byte(gitIgnore), 0644)
	}()
	// Create and replace package on serverfile
	mainServerData, _ := fs.ReadFile(mainServerFile, "server/server.go")
	// Replace "package server" with "package main"
	// To serve this lib we had to change package main on server file
	replacedpackage := []byte(strings.ReplaceAll(string(mainServerData), "package server", "package main"))
	modifiedData := []byte(strings.ReplaceAll(string(replacedpackage), "startServer()", "main()"))

	if err := os.WriteFile("main.go", modifiedData, 0644); err != nil {

		return fmt.Errorf("error creating file %s: %w", "main.go", err)
	}

	// Create Tailwind with special permissions for execution
	data, _ := fs.ReadFile(tailwindCSS, "tailwindcss")
	// Write the file with executable permissions (0755)
	if err := os.WriteFile("tailwindcss", data, 0755); err != nil {

		return fmt.Errorf("error creating file %s: %w", "tailwindcss", err)
	}

	if err := createFiles(cliFiles); err != nil {
		return err
	}
	if err := createFiles(publicFolderFiles); err != nil {
		return err
	}
	if err := createFiles(srcFiles); err != nil {
		return err
	}
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

	// Replace project name on all files
	if err := replaceOnFile("gothic-example", projectName, "samconfig.toml"); err != nil {

		return fmt.Errorf("error replacing project name to file %s: %w", "samconfig.toml", err)
	}

	if err := replaceOnFile("gothic-example", projectName, "makefile"); err != nil {

		return fmt.Errorf("error replacing project name to file %s: %w", "makefile", err)
	}

	if err := replaceOnFile("gothic-example", projectName, "template.yaml"); err != nil {

		return fmt.Errorf("error replacing project name to file %s: %w", "template.yaml", err)
	}

	return nil
}

func replaceOnFile(originalString string, replaceString string, filePath string) error {
	fileContent, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	replacedfile := []byte(strings.ReplaceAll(string(fileContent), originalString, replaceString))

	if err := os.WriteFile(filePath, replacedfile, 0644); err != nil {

		return err
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
		if err := buildAndReplace(name, indexPage, "src/pages/index.templ", "Index", "src/pages/"+name+".templ"); err != nil {
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

		if err := buildAndReplace(name, indexPage, "src/pages/index.templ", "Index", "src/pages/"+name+".templ"); err != nil {
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

		if err := buildAndReplace(name, revalidatePage, "src/pages/revalidate.templ", "Revalidate", "src/pages/"+name+".templ"); err != nil {
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

		if err := buildAndReplace(name, apiExample, "src/api/helloWorld.go", "HelloWorld", "src/api/"+name+".go"); err != nil {
			return err
		}
		templ := exec.Command("make", "templ")
		templ.Run()
		originalRouteExample := `Please add this to your api routes to use the component:

				router.Get("/", api.HelloWorld)`

		fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))

	case "isr-api-route":
		if err := buildAndReplace(name, apiExample, "src/api/helloWorld.go", "HelloWorld", "src/api/"+name+".go"); err != nil {
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
		if err := buildAndReplace(name, helloWorldExample, "src/components/helloWorld.templ", "HelloWorld", "src/components/"+name+".templ"); err != nil {
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
		if err := buildAndReplace(name, helloWorldExample, "src/components/helloWorld.templ", "HelloWorld", "src/components/"+name+".templ"); err != nil {
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
		if err := buildAndReplace(name, lazyLoadExample, "src/components/lazyLoad.templ", "LazyLoad", "src/components/"+name+".templ"); err != nil {
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
