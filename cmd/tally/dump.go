package main

import (
  "github.com/kr/pretty"
	"github.com/qmsk/e2/tally"
)

type DebugModule struct {
  Output		  bool	`long:"debug-output" description:"Dump tally state on stdout"`

  tallyChan       chan tally.State
}

func init() {
  registerModule("Debug", &DebugModule{})
}

func (module *DebugModule) run() {
  for tallyState := range module.tallyChan {
    pretty.Print(tallyState)
  }
}

func (module *DebugModule) start(t *tally.Tally) error {
  if !module.Output {
    return nil
  }

  module.tallyChan = make(chan tally.State)
  t.Register(module.tallyChan)

  go module.run()

  return nil
}

func (module *DebugModule) stop() error {
  if module.tallyChan != nil {
    close(module.tallyChan)
  }

  return nil
}
