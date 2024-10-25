package handler

import (
	"net/http"

	"github.com/a-h/templ"
)

func Render(r *http.Request, w http.ResponseWriter, component templ.Component) error {
	return component.Render(r.Context(), w)
}
