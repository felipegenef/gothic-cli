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
