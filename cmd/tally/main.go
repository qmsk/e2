package main

import (
    "github.com/qmsk/e2/client"
    "github.com/qmsk/e2/discovery"
    "github.com/jessevdk/go-flags"
    "log"
    "github.com/qmsk/e2/tally"
)

var options = struct{
    DiscoveryOptions    discovery.Options       `group:"E2 Discovery"`
    ClientOptions       client.Options          `group:"E2 XML"`
    TallyOptions        tally.Options           `group:"Tally"`
	GPIOOptions			tally.GPIOOptions		`group:"Tally GPIO"`

	GPIO	bool	`long:"gpio"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
    if _, err := parser.Parse(); err != nil {
        log.Fatalf("%v\n", err)
    }

    tally, err := options.TallyOptions.Tally(options.ClientOptions, options.DiscoveryOptions)
    if err != nil {
        log.Fatalf("Tally: %v\n", err)
    }

	if options.GPIO {
		if tallyGPIO, err := options.GPIOOptions.Make(tally); err != nil {
			log.Fatalf("Start GPIO: %v", err)
		} else {
			defer tallyGPIO.Close()
		}
	}

    if err := tally.Run(); err != nil {
        log.Fatalf("Tally.Run: %v\n", err)
    } else {
        log.Printf("Exit")
    }
}
