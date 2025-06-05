package api

import (
	"encoding/json"
	"net/http"
	 routes "github.com/felipegenef/gothic-cli/pkg/helpers/routes"
	"net/http"
)

type {{.RouteName}}Response struct {
	Message string `json:"message"`
}

var {{.RouteName}}Config = routes.ApiRouteConfig{
	HttpMethod: routes.GET,
}

func {{.RouteName}}(w http.ResponseWriter, r *http.Request) {

	response, _ := json.Marshal({{.RouteName}}Response{"Hello World from GOTH API ROUTE"})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
