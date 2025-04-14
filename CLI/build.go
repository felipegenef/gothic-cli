package cli

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type BuildCommand struct {
	cli           *GothicCli
	gothicCliData GothicCliData
}

func NewBuildCommandCli() BuildCommand {
	return BuildCommand{}
}

func (command *BuildCommand) Build(buildCmd *string, data GothicCliData) {
	name := command.promptBuildCommandName(*buildCmd)
	command.gothicCliData = data
	if name != "" {
		command.handleBuild(*buildCmd, name)
	}
}

func (command *BuildCommand) promptBuildCommandName(buildCmd string) string {
	var name string
	fmt.Printf("Enter the name for the %s (in camel case, e.g., MyPageExample,MyApiRoute,MyComponent etc...): ", buildCmd)
	fmt.Scanln(&name)

	// Validate camel case
	if matched, _ := regexp.MatchString(`^[A-Z][a-zA-Z0-9]*$`, name); !matched {
		fmt.Println("Invalid name format. Please use camel case (start with uppercase letter, followed by letters and digits).")
	}
	return name
}

func (command *BuildCommand) handleBuild(buildType, name string) error {
	command.cli.GetConfig()
	command.cli.Templates.BuildCMDTemplateInfo = BuildCMDTemplateInfo{
		PageName:      name,
		RouteName:     name,
		ComponentName: name,
		GoModName:     command.cli.config.GoModName,
	}

	switch buildType {
	case "page":
		command.buildPage(name)
	case "static-page":
		command.buildStaticPage(name)
	case "isr-page":
		command.buildISRPage(name)
	case "api-route":
		command.buildApiRoute(name)
	case "isr-api-route":
		command.buildIsrApiRoute(name)
	case "component":
		command.buildComponent(name)
	case "isr-component":
		command.buildIsrComponent(name)
	case "lazy-load-component":
		command.buildLazyLoadComponent(name)
	default:
		fmt.Println("Unknown build type. Use one of: page, static-page, isr-page, api-route, isr-api-route, component, isr-component, lazy-load-component.")
	}

	return nil
}

func (command *BuildCommand) buildPage(name string) {
	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/pages/index.templ", "src/pages/"+name+".templ", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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

func (command *BuildCommand) buildStaticPage(name string) {
	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/pages/index.templ", "src/pages/"+name+".templ", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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

func (command *BuildCommand) buildISRPage(name string) {
	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/pages/revalidate.templ", "src/pages/"+name+".templ", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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

func (command *BuildCommand) buildApiRoute(name string) {

	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/api/helloWorld.go", "src/api/"+name+".go", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
		log.Fatalf("Error generating Api Route: %w", err)
	}
	templ := exec.Command("make", "templ")
	templ.Run()
	originalRouteExample := `Please add this to your api routes to use the component:

			router.Get("/", api.HelloWorld)`

	fmt.Println(strings.ReplaceAll(originalRouteExample, "HelloWorld", name))
}

func (command *BuildCommand) buildIsrApiRoute(name string) {

	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/api/helloWorld.go", "src/api/"+name+".go", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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

func (command *BuildCommand) buildComponent(name string) {
	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/components/helloWorld.templ", "src/components/"+name+".templ", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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

func (command *BuildCommand) buildIsrComponent(name string) {
	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/components/helloWorld.templ", "src/components/"+name+".templ", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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

func (command *BuildCommand) buildLazyLoadComponent(name string) {
	if err := command.cli.Templates.CreateFromTemplate(command.gothicCliData.SrcFolder, "src/components/lazyLoad.templ", "src/components/"+name+".templ", command.cli.Templates.BuildCMDTemplateInfo); err != nil {
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
