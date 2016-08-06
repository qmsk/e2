package main

import (
	"github.com/qmsk/e2/nixie"
	"github.com/qmsk/e2/tally"
	"log"
)

type NixieModule struct {
	nixie.Options

	nixie *nixie.Nixie

	Enabled bool `long:"nixie" description:"Enable Nixie-hetec output"`
}

func init() {
	registerModule("Nixie", &NixieModule{})
}

func (module *NixieModule) start(tally *tally.Tally) error {
	if !module.Enabled {
		return nil
	}

	if nixie, err := module.Options.Make(); err != nil {
		return err
	} else {
		module.nixie = nixie
	}

	log.Printf("Nixie: Register tally")

	module.nixie.RegisterTally(tally)

	return nil
}

func (module *NixieModule) stop() error {
	if module.nixie != nil {
		module.nixie.Close()
	}

	return nil
}
