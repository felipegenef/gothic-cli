package helpers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	helpers "github.com/felipegenef/gothic-cli/pkg/helpers"
	"github.com/go-chi/chi/v5"
)

type RouteConfig[T any] struct {
	Type            ConfigType
	HttpMethod      HttpMethod
	RevalidateInSec int
	Middleware      func(w http.ResponseWriter, r *http.Request) T
}

func (config *RouteConfig[T]) RegisterRoute(r chi.Router, httpPath string, component func(T) templ.Component) {
	if config.Type == STATIC {
		switch config.HttpMethod {
		case GET:
			r.Get(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "max-age=31536000")
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case POST:
			r.Post(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "max-age=31536000")
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case PUT:
			r.Put(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "max-age=31536000")
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case PATCH:
			r.Patch(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "max-age=31536000")
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case DELETE:
			r.Delete(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "max-age=31536000")
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		}

	}

	if config.Type == DYNAMIC {
		switch config.HttpMethod {
		case GET:
			r.Get(httpPath, func(w http.ResponseWriter, r *http.Request) {
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case POST:
			r.Post(httpPath, func(w http.ResponseWriter, r *http.Request) {
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case PUT:
			r.Put(httpPath, func(w http.ResponseWriter, r *http.Request) {
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case PATCH:
			r.Patch(httpPath, func(w http.ResponseWriter, r *http.Request) {
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case DELETE:
			r.Delete(httpPath, func(w http.ResponseWriter, r *http.Request) {
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		}

	}

	if config.Type == ISR {
		switch config.HttpMethod {
		case GET:
			r.Get(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", fmt.Sprintf(
					"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
					config.RevalidateInSec, config.RevalidateInSec, config.RevalidateInSec,
				))
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case POST:
			r.Post(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", fmt.Sprintf(
					"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
					config.RevalidateInSec, config.RevalidateInSec, config.RevalidateInSec,
				))
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case PUT:
			r.Put(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", fmt.Sprintf(
					"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
					config.RevalidateInSec, config.RevalidateInSec, config.RevalidateInSec,
				))
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case PATCH:
			r.Patch(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", fmt.Sprintf(
					"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
					config.RevalidateInSec, config.RevalidateInSec, config.RevalidateInSec,
				))
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		case DELETE:
			r.Delete(httpPath, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", fmt.Sprintf(
					"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
					config.RevalidateInSec, config.RevalidateInSec, config.RevalidateInSec,
				))
				config.Render(r, w, component(config.Middleware(w, r)))
			})
		}
	}

}

func (config *RouteConfig[T]) Render(r *http.Request, w http.ResponseWriter, component templ.Component) error {
	return component.Render(r.Context(), w)
}

type ApiRouteConfig struct {
	HttpMethod HttpMethod
}

func (config *ApiRouteConfig) RegisterRoute(r chi.Router, httpPath string, fn func(w http.ResponseWriter, r *http.Request)) {
	switch config.HttpMethod {
	case GET:
		r.Get(httpPath, fn)
	case POST:
		r.Post(httpPath, fn)
	case PUT:
		r.Put(httpPath, fn)
	case PATCH:
		r.Patch(httpPath, fn)
	case DELETE:
		r.Delete(httpPath, fn)
	}

}

func (config *ApiRouteConfig) Render(r *http.Request, w http.ResponseWriter, component templ.Component) error {
	return component.Render(r.Context(), w)
}

type ConfigType int

const (
	ISR ConfigType = iota
	STATIC
	DYNAMIC
)

type HttpMethod int

const (
	GET HttpMethod = iota
	POST
	PUT
	PATCH
	DELETE
)

type RouteTemplate struct {
	FunctionName      string
	ConfigName        string
	PackageName       string
	ConfigPackageName string
	HttpPath          string
}

type Imports struct {
	Package     string
	PackagePath string
}

type TemplateInfo struct {
	GoModName string
	Imports   []Imports
	Routes    []RouteTemplate
	ApiRoutes []RouteTemplate
}

type FileBasedRouteHelper struct {
	TemplateInfo            TemplateInfo
	PackageRegex            *regexp.Regexp
	RouteConfigNameRegex    *regexp.Regexp
	ApiRouteConfigNameRegex *regexp.Regexp
	RouteFuncNameRegex      *regexp.Regexp
	ApiRouteFuncNameRegex   *regexp.Regexp
	OutputFile              string
	Template                helpers.TemplateHelper
}

func NewFileBasedRouteHelper() FileBasedRouteHelper {
	return FileBasedRouteHelper{
		OutputFile:              "./src/routes/autoGenRoutes.go",
		PackageRegex:            regexp.MustCompile(`(?m)^package\s+(\w+)`),
		RouteConfigNameRegex:    regexp.MustCompile(`(?m)^var\s+(\w+)\s*=\s*routes\.RouteConfig\[[^\]]+\]\s*{([^}]*)}`),
		ApiRouteConfigNameRegex: regexp.MustCompile(`(?m)^var\s+(\w+)\s*=\s*routes\.ApiRouteConfig\s*{([^}]+)}`),
		RouteFuncNameRegex:      regexp.MustCompile(`(?m)^func\s+(\w+)\s*\(.*\)\s+templ\.Component\s*{`),
		ApiRouteFuncNameRegex:   regexp.MustCompile(`(?m)^func\s+(\w+)\s*\(.*\)\s*{`),
		Template:                helpers.NewTemplateHelper(),
	}
}

var DefaultConfig = RouteConfig[any]{
	Type:       DYNAMIC,
	HttpMethod: GET,
	Middleware: func(w http.ResponseWriter, r *http.Request) any {
		return nil
	},
}

func (helper *FileBasedRouteHelper) Render(goModName string) error {
	fmt.Printf("Starting to read dirs...\n")

	// 1️⃣ Walk through ./src/pages
	err := filepath.Walk("./src/pages", func(path string, info os.FileInfo, err error) error {
		var route RouteTemplate
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "templ.go") {
			route.ConfigName = "DefaultConfig"
			route.ConfigPackageName = "routes"
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			packageMatch := helper.PackageRegex.FindStringSubmatch(string(content))
			if len(packageMatch) > 1 {
				route.PackageName = packageMatch[1]
				route.ConfigPackageName = packageMatch[1]
				dirName := filepath.Base(filepath.Dir(path))
				importStruct := Imports{
					Package:     route.PackageName,
					PackagePath: fmt.Sprintf("%s/src/%s", goModName, dirName),
				}
				helper.TemplateInfo.Imports = append(helper.TemplateInfo.Imports, importStruct)
			}

			configMatch := helper.RouteConfigNameRegex.FindStringSubmatch(string(content))
			if len(configMatch) > 1 {
				route.ConfigName = configMatch[1]
			} else {
				route.ConfigName = "DefaultConfig"
				route.ConfigPackageName = "routes"
			}

			funcMatch := helper.RouteFuncNameRegex.FindStringSubmatch(string(content))
			if len(funcMatch) > 1 {
				route.FunctionName = funcMatch[1]
			}

			route.HttpPath = helper.normalizeHttpPath(path)
			if route.FunctionName != "" {
				helper.TemplateInfo.Routes = append(helper.TemplateInfo.Routes, route)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk through pages: %w", err)
	}

	// 2️⃣ Walk through ./src/components
	err = filepath.Walk("./src/components", func(path string, info os.FileInfo, err error) error {
		var route RouteTemplate
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "templ.go") {
			route.ConfigName = "DefaultConfig"
			route.ConfigPackageName = "routes"
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			packageMatch := helper.PackageRegex.FindStringSubmatch(string(content))
			if len(packageMatch) > 1 {
				route.PackageName = packageMatch[1]
				route.ConfigPackageName = packageMatch[1]
				dirName := filepath.Base(filepath.Dir(path))
				importStruct := Imports{
					Package:     route.PackageName,
					PackagePath: fmt.Sprintf("%s/src/%s", goModName, dirName),
				}
				helper.TemplateInfo.Imports = append(helper.TemplateInfo.Imports, importStruct)
			}

			configMatch := helper.RouteConfigNameRegex.FindStringSubmatch(string(content))
			if len(configMatch) > 1 {
				route.ConfigName = configMatch[1]
			} else {
				route.ConfigName = "DefaultApiConfig"
				route.ConfigPackageName = "routes"
			}

			funcMatch := helper.RouteFuncNameRegex.FindStringSubmatch(string(content))
			if len(funcMatch) > 1 {
				route.FunctionName = funcMatch[1]
			}

			route.HttpPath = helper.normalizeHttpPath(path)
			if route.FunctionName != "" {
				helper.TemplateInfo.Routes = append(helper.TemplateInfo.Routes, route)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk through components: %w", err)
	}

	// 3️⃣ Walk through ./src/api
	err = filepath.Walk("./src/api", func(path string, info os.FileInfo, err error) error {
		var route RouteTemplate
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".go") {
			route.ConfigName = "DefaultApiConfig"
			route.ConfigPackageName = "routes"
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			packageMatch := helper.PackageRegex.FindStringSubmatch(string(content))
			if len(packageMatch) > 1 {
				route.PackageName = packageMatch[1]
				route.ConfigPackageName = packageMatch[1]
				dirName := filepath.Base(filepath.Dir(path))
				importStruct := Imports{
					Package:     route.PackageName,
					PackagePath: fmt.Sprintf("%s/src/%s", goModName, dirName),
				}
				helper.TemplateInfo.Imports = append(helper.TemplateInfo.Imports, importStruct)
			}

			configMatch := helper.ApiRouteConfigNameRegex.FindStringSubmatch(string(content))
			if len(configMatch) > 1 {
				route.ConfigName = configMatch[1]
			} else {
				route.ConfigName = "DefaultApiConfig"
				route.ConfigPackageName = "routes"
			}

			funcMatch := helper.ApiRouteFuncNameRegex.FindStringSubmatch(string(content))
			if len(funcMatch) > 1 {
				route.FunctionName = funcMatch[1]
			}

			route.HttpPath = helper.normalizeHttpPath(path)
			if route.FunctionName != "" {
				helper.TemplateInfo.ApiRoutes = append(helper.TemplateInfo.ApiRoutes, route)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk through api: %w", err)
	}

	// 4️⃣ Deduplicate imports
	helper.TemplateInfo.GoModName = goModName
	uniqueImports := make(map[string]Imports)
	for _, imp := range helper.TemplateInfo.Imports {
		uniqueImports[imp.PackagePath] = imp
	}

	helper.TemplateInfo.Imports = make([]Imports, 0, len(uniqueImports))
	for _, imp := range uniqueImports {
		helper.TemplateInfo.Imports = append(helper.TemplateInfo.Imports, imp)
	}

	// 5️⃣ Render template
	return helper.Template.UpdateFromTemplate("./.gothicCli/templates/autoGenRoutes.go", helper.OutputFile, helper.TemplateInfo)
}

func (helper *FileBasedRouteHelper) normalizeHttpPath(path string) string {
	// Remove the _templ.go or .go extension
	path = strings.TrimSuffix(path, "_templ.go")
	path = strings.TrimSuffix(path, ".go")

	// Remove "src/pages", "src/components", or "src" prefixes
	path = strings.TrimPrefix(path, "src/pages")
	path = strings.TrimPrefix(path, "src/components")
	path = strings.TrimPrefix(path, "src")

	// Normalize index to root or parent path
	if strings.HasSuffix(path, "/index") {
		path = strings.TrimSuffix(path, "/index")
		if path == "" {
			return "/" // root index
		}
	}

	return path
}
