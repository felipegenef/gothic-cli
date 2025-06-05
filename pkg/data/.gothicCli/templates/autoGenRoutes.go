package routes

import (
	"fmt"
	"net/http"

	routes "github.com/felipegenef/gothic-cli/pkg/helpers/routes"
	handler "{{.GoModName}}/src/utils"
	{{ range .Imports }}
	{{.Package}} "{{.PackagePath}}"
	{{ end }}

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

var FileBasedRoutes = GenerateRoutes()

type FileBasedRoutesResults struct {
	StaticRoutes  func(staticRoutes chi.Router)
	DynamicRoutes func(dynamicRoutes chi.Router)
	IsrRoutes     func(isrRoutes chi.Router)
	ApiRoutes     func(apiRoutes chi.Router)
}


type RouteInfo struct {
	Middleware func(w http.ResponseWriter, r *http.Request)
	HttpMethod string
	HttpPath   string
	Revalidate int
	Component  templ.Component
}

func GenerateRoutes() FileBasedRoutesResults {
	var result FileBasedRoutesResults

	result.StaticRoutes = func(r chi.Router) {

		{{ range .Routes }}
			if {{.ConfigPackageName}}.{{.ConfigName}}.Type ==routes.STATIC{
				switch {{.ConfigPackageName}}.{{.ConfigName}}.HttpMethod {
				case routes.GET:
					r.Get("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Cache-Control", "max-age=31536000")
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.POST:
					r.Post("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Cache-Control", "max-age=31536000")
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.PUT:
					r.Put("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Cache-Control", "max-age=31536000")
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.PATCH:
					r.Patch("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Cache-Control", "max-age=31536000")
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.DELETE:
					r.Delete("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Cache-Control", "max-age=31536000")
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				}

			}
		{{ end }}

	}

	result.DynamicRoutes = func(r chi.Router) {

		{{ range .Routes }}
			if {{.ConfigPackageName}}.{{.ConfigName}}.Type ==routes.DYNAMIC{
				switch {{.ConfigPackageName}}.{{.ConfigName}}.HttpMethod {
				case routes.GET:
					r.Get("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.POST:
					r.Post("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.PUT:
					r.Put("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.PATCH:
					r.Patch("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.DELETE:
					r.Delete("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				}

			}
		{{ end }}

	}

	result.IsrRoutes = func(r chi.Router) {

		{{ range .Routes }}
			if {{.ConfigPackageName}}.{{.ConfigName}}.Type ==routes.ISR{
				switch {{.ConfigPackageName}}.{{.ConfigName}}.HttpMethod {
				case routes.GET:
					r.Get("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
										w.Header().Set("Cache-Control", fmt.Sprintf(
						"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
						{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec, {{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,
					))
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.POST:
					r.Post("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
										w.Header().Set("Cache-Control", fmt.Sprintf(
						"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
						{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec, {{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,
					))
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.PUT:
					r.Put("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
										w.Header().Set("Cache-Control", fmt.Sprintf(
						"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
						{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec, {{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,
					))
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.PATCH:
					r.Patch("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
										w.Header().Set("Cache-Control", fmt.Sprintf(
						"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
						{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec, {{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,
					))
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				case routes.DELETE:
					r.Delete("{{.HttpPath}}", func(w http.ResponseWriter, r *http.Request) {
										w.Header().Set("Cache-Control", fmt.Sprintf(
						"max-age=%v, stale-while-revalidate=%v, stale-if-error=%v",
						{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec, {{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,{{.ConfigPackageName}}.{{.ConfigName}}.RevalidateInSec,
					))
						result := {{.ConfigPackageName}}.{{.ConfigName}}.Middleware(w, r)
						handler.Render(r, w, {{.PackageName}}.{{.FunctionName}}(result))
					})
				}

			}
		{{ end }}
	}

	result.ApiRoutes = func(r chi.Router) {
		{{ range .ApiRoutes }}
		switch {{.ConfigPackageName}}.{{.ConfigName}}.HttpMethod {
		case routes.GET:
			r.Get("{{.HttpPath}}", {{.PackageName}}.{{.FunctionName}})
		case routes.POST:
			r.Post("{{.HttpPath}}", {{.PackageName}}.{{.FunctionName}})
		case routes.PUT:
			r.Put("{{.HttpPath}}", {{.PackageName}}.{{.FunctionName}})
		case routes.PATCH:
			r.Patch("{{.HttpPath}}", {{.PackageName}}.{{.FunctionName}})
		case routes.DELETE:
			r.Delete("{{.HttpPath}}", {{.PackageName}}.{{.FunctionName}})
		}
		{{ end }}

	}

	return result

}
