package tally

import (
    "github.com/kidoman/embd"
	"fmt"
    _ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"log"
)

type GPIOOptions struct {
	TallyPins	[]string	`long:"gpio-tally-pin"`
}

func (options GPIOOptions) Make(tally *Tally) (*GPIO, error) {
	var gpio = GPIO{
		options:    options,
		TallyPins:	make(map[ID]embd.DigitalPin),
	}

	if err := gpio.init(options); err != nil {
		return nil, err
	}

	gpio.register(tally)

	return &gpio, nil
}

type GPIO struct {
	options		GPIOOptions

    TallyPins   map[ID]embd.DigitalPin

	stateChan	chan State
}

func (gpio *GPIO) init(options GPIOOptions) error {
	if err := embd.InitGPIO(); err != nil {
		return fmt.Errorf("embd.InitGPIO: %v", err)
	}

	for i, pinName := range options.TallyPins {
		if pin, err := embd.NewDigitalPin(pinName); err != nil {
			return fmt.Errorf("embd.NewDigitalPin %v: %v", pinName, err)

		// Writing as "out" defaults to initializing the value as low.
		} else if err := pin.SetDirection(embd.Out); err != nil {
			return fmt.Errorf("pin.SetDirection %v: %v", pinName, err)
		} else {
			gpio.TallyPins[ID(i + 1)] = pin
		}
	}

	return nil
}

func (gpio *GPIO) register(tally *Tally) {
	gpio.stateChan = make(chan State)

	go gpio.run()

	tally.register(gpio.stateChan)
}

func (gpio *GPIO) close() {
	for id, pin := range gpio.TallyPins {
		if err := pin.Close(); err != nil {
			log.Printf("tally:GPIO: close pin %v:%v: %v", id, pin, err)
		}
	}
}

func (gpio *GPIO) run() {
	defer gpio.close()

	for state := range gpio.stateChan {
		log.Printf("tally:GPIO: Update:")

		for id, pin := range gpio.TallyPins {
			pinValue := embd.Low

			if status, exists := state.Tally[id]; !exists {
				// missing tally state for pin
			} else if status.Program {
				log.Printf("tally:GPIO:\tpin %v:%v high: %v", id, gpio.options.TallyPins[id-1], status)

				pinValue = embd.High
			}

			if err := pin.Write(pinValue); err != nil {
				log.Printf("tally:GPIO: write pin %v:%v: %v", id, pin, err)
			}
		}
	}
}

// After Tally has stopped, close all GPIOs
func (gpio *GPIO) Close() {
	close(gpio.stateChan)
}
