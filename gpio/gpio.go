// GPIO pin output from tally state
//
// Supports two status pins and N tally pins.
//
// The green status pin is high to indicate tally activity. In the idle state (no tally sources discovered), the green
// status pin blinks on slowly. In the active state (at least one tally source connected), the green status pin is on,
// and blinks off on changes.
//
// The red status pin is normally low, and is set high when there are any tally source errors.
//
// Each tally pin corresponds to the sequentially numbered tally ID. It will be set high when the tally ID is out on
// any output program.
package gpio

import (
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/qmsk/e2/tally"
	"log"
	"sync"
	"time"
)

type Options struct {
	StatusGreenPin	string	 `long:"gpio-green-pin" value-name:"GPIO-PIN" description:"GPIO pin for green status LED"`
	StatusRedPin	string	 `long:"gpio-red-pin" value-name:"GPIO-PIN" description:"GPIO pin for red status LED"`

	TallyPins		[]string `long:"gpio-tally-pin" value-name:"GPIO-PIN" description:"Pass each tally pin as a separate option"`
}

func (options Options) Make() (*GPIO, error) {
	var gpio = GPIO{
		options:   options,
		tallyPins: make(map[tally.ID]*Pin),
	}

	if err := gpio.init(options); err != nil {
		return nil, err
	}

	return &gpio, nil
}

type GPIO struct {
	options Options

	tallyPins		map[tally.ID]*Pin

	// red pin is high if there are sources with errors
	statusRedPin	*Pin

	// green pin is high if there are sources with tallys
	statusGreenPin	*Pin

	tallyChan chan tally.State
	waitGroup sync.WaitGroup
}

func (gpio *GPIO) init(options Options) error {
	if err := embd.InitGPIO(); err != nil {
		return fmt.Errorf("embd.InitGPIO: %v", err)
	}

	for i, pinName := range options.TallyPins {
		id := tally.ID(i+1)

		if pin, err := openPin(fmt.Sprintf("tally:%d", id), pinName); err != nil {
			return err
		} else {
			gpio.tallyPins[tally.ID(i+1)] = pin
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

func (gpio *GPIO) RegisterTally(t *tally.Tally) {
	gpio.tallyChan = make(chan tally.State)
	gpio.waitGroup.Add(1)

	go gpio.run()

	t.Register(gpio.tallyChan)
}

func (gpio *GPIO) close() {
	defer gpio.waitGroup.Done()

	log.Printf("GPIO: Close pins..")

	if gpio.statusGreenPin != nil {
		gpio.statusGreenPin.Close(&gpio.waitGroup)
	}
	if gpio.statusRedPin != nil {
		gpio.statusRedPin.Close(&gpio.waitGroup)
	}

	for _, pin := range gpio.tallyPins {
		pin.Close(&gpio.waitGroup)
	}

}

func (gpio *GPIO) updateTally(state tally.State) {
	log.Printf("GPIO: Update tally State:")

	var statusGreen = false
	var statusRed = false

	for id, pin := range gpio.tallyPins {
		var pinState = false

		if status, exists := state.Tally[id]; !exists {
			// missing tally state for pin
		} else {
			statusGreen = true

			if status.Status.Program {
				log.Printf("GPIO:\ttally pin %v high: %v", pin, status)

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
		log.Printf("GPIO: status:green high: blink")

		// when connected, blink off for 100ms on every update
		gpio.statusGreenPin.Blink(false, 100 * time.Millisecond)
	} else {
		log.Printf("GPIO: status:green low: cycle")

		// when not connected, blink on for 100ms every 1s
		gpio.statusGreenPin.BlinkCycle(true, 100 * time.Millisecond, 1 * time.Second)
	}

	if gpio.statusRedPin == nil {

	} else if statusRed {
		log.Printf("GPIO: status:red blink: cycle")

		gpio.statusRedPin.BlinkCycle(true, 500 * time.Millisecond, 500 * time.Millisecond)
	} else {
		gpio.statusRedPin.Set(false)
	}
}

func (gpio *GPIO) run() {
	defer gpio.close()

	for state := range gpio.tallyChan {
		gpio.updateTally(state)
	}

	log.Printf("GPIO: Done")
}

// Close and Wait..
func (gpio *GPIO) Close() {
	log.Printf("GPIO: Close..")

	if gpio.tallyChan != nil {
		close(gpio.tallyChan)
	}

	gpio.waitGroup.Wait()
}
