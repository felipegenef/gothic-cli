package main

import (
	"embed"
	"fmt"

	gothicCli "github.com/felipegenef/gothic-cli/utils/cli"
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

var data = gothicCli.GothicCliData{
	PublicFolderAssets: map[string]embed.FS{
		"public/imageExample/blurred.jpeg":  publicFolder,
		"public/imageExample/original.jpeg": publicFolder,
		"public/favicon.ico":                publicFolder,
	},
	InitialFiles: map[string]embed.FS{
		// util files
		"src/utils/handler.go": srcFolder,
		// page files
		"src/pages/index.templ":      srcFolder,
		"src/pages/revalidate.templ": srcFolder,
		// layout files
		"src/layouts/layout.templ": srcFolder,
		// css files
		"src/css/app.css": srcFolder,
		// component files
		"src/components/helloWorld.templ":     srcFolder,
		"src/components/optimizedImage.templ": srcFolder,
		"src/components/lazyLoad.templ":       srcFolder,
		// api files
		"src/api/helloWorld.go": srcFolder,
		// root files
		"makefile":           makeFile,
		"tailwind.config.js": tailwindConfig,
		"README.md":          readme,
		"gothic-config.json": goticConfig,
		// cli files
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
	},
	InitialDirs: []string{
		// Root Dirs
		"public",
		".gothicCli",
		"src",
		"optimize",
		// Public Dirs
		"public/imageExample",
		// Cli Dirs
		".gothicCli/HotReload",
		".gothicCli/imgOptimization",
		".gothicCli/imgOptimization/setup",
		".gothicCli/buildSamTemplate",
		".gothicCli/buildSamTemplate/templates",
		".gothicCli/buildSamTemplate/cleanup",
		".gothicCli/CdnAddOrRemoveAssets",
		".gothicCli/sam",
		// Src Dirs
		"src/api",
		"src/components",
		"src/css",
		"src/layouts",
		"src/pages",
		"src/utils",
	},
	GitIgnore: gitIgnore,
	EnvSample: envs,
	Env:       envs,
	AirToml:   airToml,
	Tailwind: gothicCli.TailWindCSS{
		Mac:     tailwindCSSMac,
		Windows: tailwindCSSWindows,
		Linux:   tailwindCSSLinux,
		Config:  tailwindConfig,
	},
	GoticConfig:     goticConfig,
	Readme:          readme,
	MakeFile:        makeFile,
	SrcFolder:       srcFolder,
	PublicFolder:    publicFolder,
	ServerFolder:    serverFolder,
	GothicCliFolder: gothicCliFolder,
	CustomTemplateBasedPages: map[string]string{
		"src/pages/revalidate.templ": "Revalidate",
		"src/pages/index.templ":      "Index",
	},
	CustomTemplateBasedComponents: map[string]string{
		"src/components/helloWorld.templ": "HelloWorld",
		"src/components/lazyLoad.templ":   "LazyLoad",
	},
	CustomTemplateBasedRoutes: map[string]string{
		"src/api/helloWorld.go": "HelloWorld",
	},
}

func main() {
	cli := gothicCli.NewCli()
	commands := cli.WaitForCommands()

	if *commands.Help {
		cli.PromptHelpInstructions()
		return
	}
	if *commands.Init {
		cli.CreateNewGothicApp(data)
	} else if *commands.Build != "" {
		cli.BuildCommand(commands.Build, data)
	} else {
		fmt.Println("Use --init to initialize the project or --build to build a boilerplate example.")
	}

}
