package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/tally"
	"github.com/qmsk/e2/web"
	"log"
	"os"
	"os/signal"
)

type module interface {
	start(tally *tally.Tally) error
	stop() error
}

var Options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 XML"`
	TallyOptions     tally.Options     `group:"Tally"`
	WebOptions       web.Options       `group:"Web API"`

	modules map[string]module
}{
	modules: make(map[string]module),
}

func registerModule(name string, module module) {
	if _, err := parser.AddGroup(name, name, module); err != nil {
		panic(err)
	}

	Options.modules[name] = module
}

var parser = flags.NewParser(&Options, flags.Default)

func start(tally *tally.Tally) {
	for moduleName, module := range Options.modules {
		if err := module.start(tally); err != nil {
			log.Fatalf("%v main: %v", moduleName, err)
		}
	}
}

func stop() {
	for moduleName, module := range Options.modules {
		if err := module.stop(); err != nil {
			log.Printf("Stop %v: %v", moduleName, err)
		}
	}
}

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}

	tally, err := Options.TallyOptions.Tally(Options.ClientOptions, Options.DiscoveryOptions)
	if err != nil {
		log.Fatalf("Tally: %v\n", err)
	}

	// setup modules
	start(tally)
	defer stop()

	// stopping
	stopChan := make(chan os.Signal)

	signal.Notify(stopChan, os.Interrupt)

	go func() {
		<-stopChan

		// second SIGINT kills
		signal.Stop(stopChan)

		tally.Stop()

	}()

	// Web
	webAPI := tally.WebAPI()
	webEvents := tally.WebEvents()

	go Options.WebOptions.Server(
		web.RoutePrefix("/api/", webAPI),
		web.RoutePrefix("/events", webEvents),
		Options.WebOptions.RouteStatic("/static/"),
		Options.WebOptions.RouteFile("/", "tally.html"),
	)

	// run
	if err := tally.Run(); err != nil {
		log.Fatalf("Tally.Run: %v\n", err)
	} else {
		log.Printf("Exit")
	}
}
