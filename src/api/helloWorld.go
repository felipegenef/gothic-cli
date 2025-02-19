package api

import (
	"encoding/json"
	"net/http"
)

type {{.RouteName}}Response struct {
	Message string `json:"message"`
}

func {{.RouteName}}(w http.ResponseWriter, r *http.Request) {

	response, _ := json.Marshal({{.RouteName}}Response{"Hello World from GOTH API ROUTE"})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
