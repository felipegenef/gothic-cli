package routes

import (

	routes "github.com/felipegenef/gothic-cli/pkg/helpers/routes"
	{{ range .Imports }}
	{{.Package}} "{{.PackagePath}}"
	{{ end }}

	"github.com/go-chi/chi/v5"
)


func RegisterFileBasedRoutes(r chi.Router) {
	{{ range .Routes }}
		{{.ConfigPackageName}}.{{.ConfigName}}.RegisterRoute(r,"{{.HttpPath}}",{{.PackageName}}.{{.FunctionName}})
	{{ end }}
	{{ range .ApiRoutes }}
		{{.ConfigPackageName}}.{{.ConfigName}}.RegisterRoute(r,"{{.HttpPath}}",{{.PackageName}}.{{.FunctionName}})
	{{ end }}

}
