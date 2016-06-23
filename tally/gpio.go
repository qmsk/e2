package tally

import (
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"log"
	"sync"
)

type GPIOOptions struct {
	StatusGreenPin	string	 `long:"gpio-green-pin"`
	StatusRedPin	string	 `long:"gpio-red-pin"`

	TallyPins		[]string `long:"gpio-tally-pin"`
}

func (options GPIOOptions) Make(tally *Tally) (*GPIO, error) {
	var gpio = GPIO{
		options:   options,
		tallyPins: make(map[ID]embd.DigitalPin),
	}

	if err := gpio.init(options); err != nil {
		return nil, err
	}

	gpio.register(tally)

	return &gpio, nil
}

type GPIO struct {
	options GPIOOptions

	tallyPins		map[ID]embd.DigitalPin

	// red pin is high if there are sources with errors
	statusRedPin	embd.DigitalPin

	// green pin is high if there are sources with tallys
	statusGreenPin	embd.DigitalPin

	stateChan chan State
	waitGroup sync.WaitGroup
}

func openPinOut(pinName string) (embd.DigitalPin, error) {
	if pin, err := embd.NewDigitalPin(pinName); err != nil {
		return nil, fmt.Errorf("embd.NewDigitalPin %v: %v", pinName, err)

		// Writing as "out" defaults to initializing the value as low.
	} else if err := pin.SetDirection(embd.Out); err != nil {
		return nil, fmt.Errorf("pin.SetDirection %v: %v", pinName, err)
	} else {
		return pin, nil
	}
}

func (gpio *GPIO) init(options GPIOOptions) error {
	if err := embd.InitGPIO(); err != nil {
		return fmt.Errorf("embd.InitGPIO: %v", err)
	}

	for i, pinName := range options.TallyPins {
		if pin, err := openPinOut(pinName); err != nil {
			return err
		} else {
			gpio.tallyPins[ID(i+1)] = pin
		}
	}

	if options.StatusGreenPin == "" {

	} else if pin, err := openPinOut(options.StatusGreenPin); err != nil {
		return err
	} else {
		gpio.statusGreenPin = pin
	}
	if options.StatusRedPin == "" {

	} else if pin, err := openPinOut(options.StatusRedPin); err != nil {
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

	for id, pin := range gpio.tallyPins {
		if err := pin.Close(); err != nil {
			log.Printf("tally:GPIO: close pin %v:%v: %v", id, pin, err)
		} else {
			closed++
		}
	}

	log.Printf("tally:GPIO: Closed %d pins", closed)
}

func (gpio *GPIO) update(state State) {
	log.Printf("tally:GPIO: Update:")

	var statusGreenValue = embd.Low
	var statusRedValue = embd.Low

	for id, pin := range gpio.tallyPins {
		pinValue := embd.Low

		if status, exists := state.Tally[id]; !exists {
			// missing tally state for pin
		} else {
			statusGreenValue = embd.High

			if status.Program {
				log.Printf("tally:GPIO:\tpin %v:%v high: %v", id, gpio.options.TallyPins[id-1], status)

				pinValue = embd.High
			}
		}

		if err := pin.Write(pinValue); err != nil {
			log.Printf("tally:GPIO: write pin %v:%v: %v", id, pin, err)
		}
	}

	if len(state.Errors) > 0 {
		statusRedValue = embd.High
	}

	// update status leds
	if gpio.statusGreenPin == nil {

	} else if err := gpio.statusGreenPin.Write(statusGreenValue); err != nil {
		log.Printf("tally:GPIO: write pin status-green:%v: %v", gpio.options.StatusGreenPin, err)
	}

	if gpio.statusRedPin == nil {

	} else if err := gpio.statusRedPin.Write(statusRedValue); err != nil {
		log.Printf("tally:GPIO: write pin status-red:%v: %v", gpio.options.StatusRedPin, err)
	}
}

func (gpio *GPIO) run() {
	defer gpio.close()

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
