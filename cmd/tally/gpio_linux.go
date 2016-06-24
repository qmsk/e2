package main

import (
	"github.com/qmsk/e2/gpio"
	"github.com/qmsk/e2/tally"
	"log"
)

type GPIOModule struct {
	gpio.Options

	gpio		*gpio.GPIO

	Enabled		bool	`long:"gpio" description:"Enable GPIO output"`
}

func init() {
	registerModule("GPIO", &GPIOModule{})
}

func (module *GPIOModule) start(tally *tally.Tally) error {
	if !module.Enabled {
		return nil
	}

	if gpio, err := module.Options.Make(); err != nil {
		return err
	} else {
		module.gpio = gpio
	}

	log.Printf("GPIO: Register tally")

	module.gpio.RegisterTally(tally)

	return nil
}

func (module *GPIOModule) stop() error {
	if module.gpio != nil {
		module.gpio.Close()
	}

	return nil
}
