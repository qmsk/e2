package main

import (
	"github.com/qmsk/e2/spiled"
	"github.com/qmsk/e2/tally"
	"log"
)

type SPILEDModule struct {
	spiled.Options

	spiled *spiled.SPILED

	Enabled bool `long:"spiled" description:"Enable SPI-LED output"`
}

func init() {
	registerModule("SPI-LED", &SPILEDModule{})
}

func (module *SPILEDModule) start(tally *tally.Tally) error {
	if !module.Enabled {
		return nil
	}

	if spiled, err := module.Options.Make(); err != nil {
		return err
	} else {
		module.spiled = spiled
	}

	log.Printf("SPI-LED: Register tally")

	module.spiled.RegisterTally(tally)

	return nil
}

func (module *SPILEDModule) stop() error {
	if module.spiled != nil {
		module.spiled.Close()
	}

	return nil
}
