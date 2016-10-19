package main

import (
	"fmt"
	"log"

	"github.com/qmsk/e2/tally"
	"github.com/qmsk/e2/universe"
)

type UniverseModule struct {
	universe.TallyOptions

	tallyDriver *universe.TallyDriver
}

func init() {
	registerModule("Universe", &UniverseModule{})
}

func (module *UniverseModule) start(tally *tally.Tally) error {
	if !module.TallyOptions.Enabled() {
		return nil
	}

	if tallyDriver, err := module.TallyOptions.TallyDriver(); err != nil {
		return fmt.Errorf("universe:TallyDriver: %v", err)
	} else {
		log.Printf("Universe:TallyDriver: %v", tallyDriver)

		module.tallyDriver = tallyDriver
	}

	module.tallyDriver.RegisterTally(tally)

	return nil
}

func (module *UniverseModule) stop() error {
	if module.tallyDriver != nil {
		module.tallyDriver.Close()
	}

	return nil
}
