package tally

import (
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"log"
	"sync"
	"time"
)

type GPIOOptions struct {
	StatusGreenPin	string	 `long:"gpio-green-pin"`
	StatusRedPin	string	 `long:"gpio-red-pin"`

	TallyPins		[]string `long:"gpio-tally-pin"`
}

func (options GPIOOptions) Make(tally *Tally) (*GPIO, error) {
	var gpio = GPIO{
		options:   options,
		tallyPins: make(map[ID]*GPIOPin),
	}

	if err := gpio.init(options); err != nil {
		return nil, err
	}

	gpio.register(tally)

	return &gpio, nil
}

type GPIO struct {
	options GPIOOptions

	tallyPins		map[ID]*GPIOPin

	// red pin is high if there are sources with errors
	statusRedPin	*GPIOPin

	// green pin is high if there are sources with tallys
	statusGreenPin	*GPIOPin

	stateChan chan State
	waitGroup sync.WaitGroup
}

func (gpio *GPIO) init(options GPIOOptions) error {
	if err := embd.InitGPIO(); err != nil {
		return fmt.Errorf("embd.InitGPIO: %v", err)
	}

	for i, pinName := range options.TallyPins {
		id := ID(i+1)

		if pin, err := openPin(fmt.Sprintf("tally:%d", id), pinName); err != nil {
			return err
		} else {
			gpio.tallyPins[ID(i+1)] = pin
		}
	}

	if options.StatusGreenPin == "" {

	} else if pin, err := openPin("status:green", options.StatusGreenPin); err != nil {
		return err
	} else {
		gpio.statusGreenPin = pin
	}

	if options.StatusRedPin == "" {

	} else if pin, err := openPin("status:red", options.StatusRedPin); err != nil {
		return err
	} else {
		gpio.statusRedPin = pin
	}

	return nil
}

func (gpio *GPIO) register(tally *Tally) {
	gpio.stateChan = make(chan State)
	gpio.waitGroup.Add(1)

	go gpio.run()

	tally.register(gpio.stateChan)
}

func (gpio *GPIO) close() {
	defer gpio.waitGroup.Done()

	var closed = 0

	for _, pin := range gpio.tallyPins {
		if err := pin.Close(); err != nil {
			log.Printf("tally:GPIO: close pin %v: %v", pin, err)
		} else {
			closed++
		}
	}

	log.Printf("tally:GPIO: Closed %d pins", closed)
}

func (gpio *GPIO) update(state State) {
	log.Printf("tally:GPIO: Update:")

	var statusGreen = false
	var statusRed = false

	for id, pin := range gpio.tallyPins {
		var pinState = false

		if status, exists := state.Tally[id]; !exists {
			// missing tally state for pin
		} else {
			statusGreen = true

			if status.Program {
				log.Printf("tally:GPIO:\tpin %v high: %v", pin, status)

				pinState = true
			}
		}

		pin.Set(pinState)
	}

	if len(state.Errors) > 0 {
		statusRed = true
	}

	// update status leds
	if gpio.statusGreenPin == nil {

	} else if statusGreen {
		// when connected, blink off for 100ms on every update
		gpio.statusGreenPin.Blink(false, 100 * time.Millisecond)
	} else {
		// when not connected, blink on for 100ms every 1s
		gpio.statusGreenPin.BlinkCycle(true, 100 * time.Millisecond, 1 * time.Second)
	}

	if gpio.statusRedPin == nil {

	} else {
		gpio.statusRedPin.Set(statusRed)
	}
}

func (gpio *GPIO) run() {
	defer gpio.close()

	// initial
	gpio.update(State{})

	for state := range gpio.stateChan {
		gpio.update(state)
	}

	log.Printf("tally:GPIO: End")
}

// Close and Wait
func (gpio *GPIO) Close() {
	log.Printf("tally:GPIO: Closing..")

	close(gpio.stateChan)

	gpio.waitGroup.Wait()
}
