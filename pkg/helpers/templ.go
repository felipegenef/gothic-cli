package helpers

import (
	"context"
	"os"

	templ "github.com/a-h/templ/cmd/templ/generatecmd"
)

type TemplHelper struct {
}

func NewTemplHelper() TemplHelper {
	return TemplHelper{}
}

func (t *TemplHelper) Render() error {
	logger := NewLogger("error", false, os.Stdout)

	err := templ.Run(context.Background(), logger, templ.Arguments{})
	if err != nil {
		return err
	}
	return nil
}

func (t *TemplHelper) Watch() error {
	go func() {
		logger := NewLogger("error", false, os.Stdout)

		templ.Run(context.Background(), logger, templ.Arguments{
			Watch: true,
			Proxy: "http://localhost:8080",
		})
	}()
	return nil
}
