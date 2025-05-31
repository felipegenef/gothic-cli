package helpers

import (
	"log"

	"github.com/air-verse/air/runner"
)

type AirHelper struct {
}

func NewAirHelper() AirHelper {
	return AirHelper{}
}

func (a *AirHelper) Watch() error {
	go func() {

		cfg, err := runner.InitConfig("")
		if err != nil {
			log.Fatal(err)
			return
		}
		r, err := runner.NewEngineWithConfig(cfg, false)
		if err != nil {
			log.Fatal(err)
			return
		}
		r.Run()
	}()
	// TODO check for api UP and return error if timed out
	// Check if necessary or if just need to wath over templ proxy
	return nil
}
