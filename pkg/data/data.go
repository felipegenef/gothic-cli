package data

import (
	"embed"
)

//go:embed .gothicCli/templates
var templatesFolder embed.FS

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

var env string = `HTTP_LISTEN_ADDR: ":8080"
LOCAL_SERVE: "true"`

var gitIgnore string = `.env
bin
*_templ.go*
*templ.txt
node_modules
.aws-sam
tmp
optimize/*
public/styles.css
template.yaml
samconfig.toml
Dockerfile`

type TailWindCSS struct {
	Mac     embed.FS
	Windows embed.FS
	Linux   embed.FS
	Config  embed.FS
}

type GothicCliData struct {
	TemplateFiles                 map[string]embed.FS
	InitialFiles                  map[string]embed.FS
	PublicFolderAssets            map[string]embed.FS
	InitialDirs                   []string
	CustomTemplateBasedPages      map[string]string
	CustomTemplateBasedComponents map[string]string
	CustomTemplateBasedRoutes     map[string]string
	GitIgnore                     string
	Env                           string
	Tailwind                      TailWindCSS
	GoticConfig                   embed.FS
	Readme                        embed.FS
	MakeFile                      embed.FS
	SrcFolder                     embed.FS
	PublicFolder                  embed.FS
	ServerFolder                  embed.FS
	ProjectName                   string
	GoModName                     string
}

var DefaultCLIData = GothicCliData{
	PublicFolderAssets: map[string]embed.FS{
		"public/imageExample/blurred.jpeg":  publicFolder,
		"public/imageExample/original.jpeg": publicFolder,
		"public/favicon.ico":                publicFolder,
		"public/styles.css":                 publicFolder,
	},
	TemplateFiles: map[string]embed.FS{
		".gothicCli/templates/Dockerfile-template":                  templatesFolder,
		".gothicCli/templates/samconfig-template.toml":              templatesFolder,
		".gothicCli/templates/template-custom-domain-with-arn.yaml": templatesFolder,
		".gothicCli/templates/template-custom-domain.yaml":          templatesFolder,
		".gothicCli/templates/template-default.yaml":                templatesFolder,
		".gothicCli/templates/autoGenRoutes.go":                     templatesFolder,
	},
	InitialFiles: map[string]embed.FS{
		// route files
		"src/routes/autoGenRoutes.go": srcFolder,
		// page files
		"src/pages/index.templ":      srcFolder,
		"src/pages/revalidate.templ": srcFolder,
		// layout files
		"src/layouts/layout.templ": srcFolder,
		// css files
		"src/css/app.css": srcFolder,
		// component files
		"src/components/helloWorld.templ":                            srcFolder,
		"src/components/optimizedImage/var_name/var_extension.templ": srcFolder,
		"src/components/lazyLoad.templ":                              srcFolder,
		// api files
		"src/api/helloWorld.go": srcFolder,
		// root files
		"makefile":           makeFile,
		"tailwind.config.js": tailwindConfig,
		"README.md":          readme,
		"gothic-config.json": goticConfig,
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
		".gothicCli/templates",
		// Src Dirs
		"src/api",
		"src/components",
		"src/components/optimizedImage",
		"src/components/optimizedImage/var_name",
		"src/css",
		"src/layouts",
		"src/pages",
		"src/routes",
	},
	GitIgnore: gitIgnore,
	Env:       env,
	Tailwind: TailWindCSS{
		Mac:     tailwindCSSMac,
		Windows: tailwindCSSWindows,
		Linux:   tailwindCSSLinux,
		Config:  tailwindConfig,
	},
	GoticConfig:  goticConfig,
	Readme:       readme,
	MakeFile:     makeFile,
	SrcFolder:    srcFolder,
	PublicFolder: publicFolder,
	ServerFolder: serverFolder,
	CustomTemplateBasedPages: map[string]string{
		"src/pages/revalidate.templ": "Revalidate",
		"src/pages/index.templ":      "Index",
	},
	CustomTemplateBasedComponents: map[string]string{
		"src/components/helloWorld.templ":                            "HelloWorld",
		"src/components/lazyLoad.templ":                              "LazyLoad",
		"src/components/optimizedImage/var_name/var_extension.templ": "OptimizedImage",
	},
	CustomTemplateBasedRoutes: map[string]string{
		"src/api/helloWorld.go": "HelloWorld",
	},
}
