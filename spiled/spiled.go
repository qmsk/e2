// Support for SPI-based RGB LED chains (APA-102).
//
// The first LED is the status LED:
//	blue: idle
//  green: connected
//  red: errors
//  orange: connected+errors
//
// The remaining LEDs are tally LEDs, using the sequential ID numbering
//  blue: found
//  green: preview
//  red: program
//  orange: program+preview
package spiled

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/qmsk/e2/tally"
	"log"
	"strings"
	"sync"
	"time"
)

type Protocol string

const APA102	Protocol	= "apa102"
const APA102X				= "apa102x"

type Options struct {
	Channel		byte	`long:"spiled-channel" metavar:"N" description:"/dev/spidev0.N"`
	Speed		int		`long:"spiled-speed" metavar:"HZ"`
	Protocol	string	`long:"spiled-protocol" metavar:"apa102|apa102x" description:"Type of LED"`
	Count		uint	`long:"spiled-count" metavar:"COUNT" description:"Number of LEDs"`
	Debug		bool	`long:"spiled-debug" description:"Dump SPI output"`
	Intensity	uint8	`long:"spiled-intensity" metavar:"0-255" default:"255"`
	Refresh		float64 `long:"spiled-refresh" metavar:"HZ" default:"10"`

	TallyIdle		LED		`long:"spiled-tally-idle"    metavar:"RRGGBB" default:"000010"`
	TallyPreview	LED		`long:"spiled-tally-preview" metavar:"RRGGBB" default:"00ff00"`
	TallyProgram	LED		`long:"spiled-tally-program" metavar:"RRGGBB" default:"ff0000"`
	TallyBoth		LED		`long:"spiled-tally-both"    metavar:"RRGGBB" default:"ff4000"`

	StatusIdle		LED		`long:"spiled-status-idle"    metavar:"RRGGBB" default:"0000ff"`
	StatusOK		LED		`long:"spiled-status-ok"      metavar:"RRGGBB" default:"00ff00"`
	StatusWarn	    LED		`long:"spiled-status-warn"    metavar:"RRGGBB" default:"ffff00"`
	StatusError		LED		`long:"spiled-status-error"   metavar:"RRGGBB" default:"ff0000"`
}

func (options Options) Make() (*SPILED, error) {
	var spiled = SPILED{
		options:   options,
	}

	if err := spiled.init(options); err != nil {
		return nil, err
	}

	return &spiled, nil
}

type SPILED struct {
	options     Options
	protocol	Protocol
	count		uint

	spiBus	embd.SPIBus

	leds		[]LED

	tallyChan chan tally.State
	waitGroup sync.WaitGroup
}

func (spiled *SPILED) init(options Options) error {
	if err := embd.InitSPI(); err != nil {
		return fmt.Errorf("embd.InitSPI: %v", err)
	}

	// SPI
	var spiMode byte = embd.SPIMode0
	var spiChannel byte = options.Channel // /dev/spidev0.X
	var spiSpeed int = options.Speed // Hz
	var spiBitsPerWord int = 8 // bits
	var spiDelay int = 0 // us?

	spiled.protocol = Protocol(strings.ToLower(options.Protocol))
	spiled.count = options.Count

	switch spiled.protocol {
	case APA102, APA102X:
		spiMode = embd.SPIMode0
		spiBitsPerWord = 8

	default:
		return fmt.Errorf("Invalid --spiled-protocol=%v", options.Protocol)
	}

	spiled.spiBus = embd.NewSPIBus(spiMode, spiChannel, spiSpeed, spiBitsPerWord, spiDelay)

	// initial output
	spiled.leds = make([]LED, spiled.count)

	if err := spiled.write(spiled.leds, time.Time{}); err != nil {
		return err
	}

	log.Printf("SPI-LED: Open %v with %d %s LEDs", spiled.spiBus, options.Count, spiled.protocol)

	return nil
}

func (spiled *SPILED) write(leds []LED, renderTime time.Time) error {
	var packet bytes.Buffer

	var stopByte = []byte{0xff}
	var stopCount = 4 * (1 + len(leds) / 32) // one bit per byte, in frames of 32 bits

	switch spiled.protocol {
	case APA102X:
		// variation where the stop frame must be 0x00
		stopByte = []byte{0x00}
	}

	// start
	var startFrame = []byte{0x00, 0x00, 0x00, 0x00}

	packet.Write(startFrame)

	// data
	for _, led := range leds {
		var ledFrame = led.render(renderTime)

		packet.Write(ledFrame)
	}

	// stop
	var stopFrame = bytes.Repeat(stopByte, stopCount)

	packet.Write(stopFrame)

	if spiled.options.Debug {
		fmt.Println(hex.Dump(packet.Bytes()))
	}

	// send
	if _, err := spiled.spiBus.Write(packet.Bytes()); err != nil {
		return fmt.Errorf("SPIBUs.Write %v: %v", packet.Len(), err)
	} else {
		return nil
	}
}

func (spiled *SPILED) close() {
	defer spiled.waitGroup.Done()

	log.Printf("SPI-LED: Flush and close...")

	// flush empty output
	leds := make([]LED, spiled.count)

	if err := spiled.write(leds, time.Time{}); err != nil {
		log.Printf("SPILED.close: embd.SPIBus.Write: %v", err)
	}

	if err := spiled.spiBus.Close(); err != nil {
		log.Printf("SPILED.close: embd.SPIBus.Close: %v", err)
	}
}

func (spiled *SPILED) updateTally(tallyState tally.State) {
	log.Printf("SPI-LED: Update tally State:")

	leds := make([]LED, spiled.count)

	var found, errors int

	for i, led := range leds {
		if i == 0 {
			// skip status
			continue
		}

		id := tally.ID(i)

		if tally, exists := tallyState.Tally[id]; !exists {
			// missing tally state for pin
		} else {
			found++

			if tally.Status.Program && tally.Status.Preview && tally.Status.Active {
				led = spiled.options.TallyBoth
			} else if tally.Status.Preview && tally.Status.Active {
				led = spiled.options.TallyPreview
			} else if tally.Status.Program {
				led = spiled.options.TallyProgram
			} else {
				led = spiled.options.TallyIdle
			}

			led.Intensity = spiled.options.Intensity

			if tally.Errors != nil {
				led.Strobe(1 * time.Second)
			}

			log.Printf("SPI-LED %v: id=%v status=%v errors=%v led=%v", i, id, tally.Status, len(tally.Errors), led)

		}

		leds[i] = led
	}

	errors = len(tallyState.Errors)

	// status LED
	var statusLED LED

	if found > 0 && errors > 0 {
		statusLED = spiled.options.StatusWarn
	} else if errors > 0 {
		statusLED = spiled.options.StatusError
	} else if found > 0 {
		statusLED = spiled.options.StatusOK
	} else {
		statusLED = spiled.options.StatusIdle
	}

	statusLED.Intensity = spiled.options.Intensity

	log.Printf("SPI-LED: found=%v errors=%v led=%v", found, errors, statusLED)

	leds[0] = statusLED

	// refresh
	spiled.leds = leds
}

func (spiled *SPILED) run() {
	defer spiled.close()

	refreshTimer := time.Tick(time.Duration(1.0 / spiled.options.Refresh * float64(time.Second)))

	for {
		select {
		case tallyState, ok := <-spiled.tallyChan:
			if ok {
				spiled.updateTally(tallyState)
			} else {
				return
			}
		case refreshTime := <-refreshTimer:
			if err := spiled.write(spiled.leds, refreshTime); err != nil {
				log.Printf("SPI-LED: Write error: %v", err)
			}
		}
	}

	log.Printf("SPI-LED: Done")
}

func (spiled *SPILED) RegisterTally(t *tally.Tally) {
	spiled.tallyChan = make(chan tally.State)
	spiled.waitGroup.Add(1)

	go spiled.run()

	t.Register(spiled.tallyChan)
}

// Close and Wait..
func (spiled *SPILED) Close() {
	log.Printf("SPI-LED: Close..")

	if spiled.tallyChan != nil {
		close(spiled.tallyChan)
	}

	spiled.waitGroup.Wait()
}
