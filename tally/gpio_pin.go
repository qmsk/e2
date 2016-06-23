package tally

import (
	"github.com/kidoman/embd"
	"fmt"
	"time"
	"log"
)

type gpioState struct {
	value	bool
	blink	time.Duration
	cycle	time.Duration
}

// Output pin
type GPIOPin struct {
	label    string
	name	 string
	embdPin	 embd.DigitalPin

	c		chan gpioState
}

func (gp *GPIOPin) String() string {
	return fmt.Sprintf("%v=%v", gp.label, gp.name)
}

func openPin(label string, name string) (*GPIOPin, error) {
	var gp = GPIOPin{
		label: label,
		name: name,
	}

	if embdPin, err := embd.NewDigitalPin(name); err != nil {
		return nil, fmt.Errorf("embd.NewDigitalPin %v: %v", name, err)

		// Writing as "out" defaults to initializing the value as low.
	} else if err := embdPin.SetDirection(embd.Out); err != nil {
		return nil, fmt.Errorf("pin.SetDirection %v: %v", name, err)
	} else {
		gp.embdPin = embdPin
	}

	gp.c = make(chan gpioState)

	go gp.run()

	return &gp, nil
}

func (gp *GPIOPin) write(value bool) error {
	embdValue := embd.Low

	if value {
		embdValue = embd.High
	}

	return gp.embdPin.Write(embdValue)
}


func (gp *GPIOPin) run() {
	var state gpioState
	var timerChan <-chan time.Time

	for {
		select {
		case setState := <-gp.c:
			state = setState

		case <-timerChan:
			state.value = !state.value

			if state.cycle == 0 {
				state.blink = 0
			} else {
				blink := state.blink

				state.blink = state.cycle
				state.cycle = blink
			}
		}

		log.Printf("GPIOPin %v: value=%v blink=%v", gp, state.value, state.blink)

		gp.write(state.value)

		if state.blink == 0 {
			timerChan = nil
		} else {
			timerChan = time.After(state.blink)
		}
	}
}

func (gp *GPIOPin) Set(value bool) {
	gp.c <- gpioState{value:value}
}

// Set to value, and return to !value
func (gp *GPIOPin) Blink(value bool, blink time.Duration) {
	gp.c <- gpioState{value:value, blink:blink}
}

// Set to value, and cycle between !value and value
func (gp *GPIOPin) Cycle(value bool, cycle time.Duration) {
	gp.c <- gpioState{value:value, blink:cycle, cycle:cycle}
}

func (gp *GPIOPin) BlinkCycle(value bool, blink time.Duration, cycle time.Duration) {
	gp.c <- gpioState{value:value, blink:blink, cycle:cycle}
}

func (gp *GPIOPin) Close() error {
	return gp.embdPin.Close()
}
