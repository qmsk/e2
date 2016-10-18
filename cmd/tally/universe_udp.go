package main

import (
  "github.com/qmsk/e2/universe"
	"github.com/qmsk/e2/tally"
  "fmt"
  "log"
)

type UniverseModule struct {
  universe.UDPOptions

  udpTally *universe.UDPTally
}

func init() {
  registerModule("Universe", &UniverseModule{})
}

func (module *UniverseModule) start(tally *tally.Tally) error {
  if !module.UDPOptions.Enabled() {
    return nil
  }

  if udpTally, err := module.UDPOptions.UDPTally(); err != nil {
    return fmt.Errorf("universe:UDPTally: %v", err)
  } else {
    log.Printf("Universe UDPTally: %v", udpTally)

    module.udpTally = udpTally
  }

  module.udpTally.RegisterTally(tally)

  return nil
}

func (module *UniverseModule) stop() error {
  if module.udpTally != nil {
    module.udpTally.Close()
  }

  return nil
}
