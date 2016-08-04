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

	tallyChan chan tally.State
	waitGroup sync.WaitGroup
}

func (spiled *SPILED) init(options Options) error {
	if err := embd.InitSPI(); err != nil {
		return fmt.Errorf("embd.InitSPI: %v", err)
	}

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
	leds := make([]LED, spiled.count)

	for i, _ := range leds {
		leds[i] = LED{0xff, 0x00, 0x00, 0xff}
	}

	if err := spiled.write(leds); err != nil {
		return err
	}

	log.Printf("SPI-LED: Open %v with %d %s LEDs", spiled.spiBus, options.Count, spiled.protocol)

	return nil
}

func (spiled *SPILED) write(leds []LED) error {
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
		packet.Write(led.Bytes())
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

	log.Printf("SPI-LED: Close...")

	if err := spiled.spiBus.Close(); err != nil {
		log.Printf("embd.SPIBus.Close: %v", err)
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

			led.Intensity = 0xff

			if tally.Status.Program && tally.Status.Preview {
				led.Red = 0xff
				led.Green = 0x20
			} else if tally.Status.Preview {
				led.Green = 0xff
			} else if tally.Status.Program {
				led.Red = 0xff
			} else {
				led.Blue = 0x80
			}

			log.Printf("SPI-LED %v: id=%v status=%v led=%v", i, id, tally.Status, led)

		}

		leds[i] = led
	}

	errors = len(tallyState.Errors)

	// status LED
	var statusLED = LED{Intensity: 0xff}

	if found > 0 && errors > 0{
		statusLED.Red = 0xff
		statusLED.Green = 0xff
	} else if errors > 0 {
		statusLED.Red = 0xff
	} else if found > 0 {
		statusLED.Green = 0xff
	} else {
		statusLED.Blue = 0xff
	}

	log.Printf("SPI-LED: found=%v errors=%v led=%v", found, errors, statusLED)

	leds[0] = statusLED

	// write
	if err := spiled.write(leds); err != nil {
		log.Printf("SPI-LED: Write error: %v", err)
	}
}

func (spiled *SPILED) run() {
	defer spiled.close()

	for state := range spiled.tallyChan {
		spiled.updateTally(state)
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
