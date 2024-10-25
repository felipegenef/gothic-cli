package api

import (
	"encoding/json"
	"net/http"
)

type helloWorldResponse struct {
	Message string `json:"message"`
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {

	response, _ := json.Marshal(helloWorldResponse{"Hello World from GOTH API ROUTE"})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
