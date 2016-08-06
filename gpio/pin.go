package gpio

import (
	"fmt"
	"github.com/kidoman/embd"
	"sync"
	"time"
)

type pinState struct {
	value bool
	blink time.Duration
	cycle time.Duration
}

// Output pin
type Pin struct {
	label   string
	name    string
	embdPin embd.DigitalPin

	c       chan pinState
	closeWg *sync.WaitGroup
}

func (gp *Pin) String() string {
	return fmt.Sprintf("%v=%v", gp.label, gp.name)
}

func openPin(label string, name string) (*Pin, error) {
	var gp = Pin{
		label: label,
		name:  name,
	}

	if embdPin, err := embd.NewDigitalPin(name); err != nil {
		return nil, fmt.Errorf("embd.NewDigitalPin %v: %v", name, err)

		// Writing as "out" defaults to initializing the value as low.
	} else if err := embdPin.SetDirection(embd.Out); err != nil {
		return nil, fmt.Errorf("pin.SetDirection %v: %v", name, err)
	} else {
		gp.embdPin = embdPin
	}

	gp.c = make(chan pinState)

	go gp.run()

	return &gp, nil
}

func (gp *Pin) write(value bool) error {
	embdValue := embd.Low

	if value {
		embdValue = embd.High
	}

	return gp.embdPin.Write(embdValue)
}

func (gp *Pin) close() {
	gp.embdPin.Close()

	if gp.closeWg != nil {
		gp.closeWg.Done()
	}
}

func (gp *Pin) run() {
	defer gp.close()

	var state pinState
	var timerChan <-chan time.Time

	for {
		select {
		case setState, valid := <-gp.c:
			if valid {
				state = setState
			} else {
				// Close()
				return
			}

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

		// log.Printf("Pin %v: value=%v blink=%v", gp, state.value, state.blink)

		gp.write(state.value)

		if state.blink == 0 {
			timerChan = nil
		} else {
			timerChan = time.After(state.blink)
		}
	}
}

func (gp *Pin) Set(value bool) {
	gp.c <- pinState{value: value}
}

// Set to value, and return to !value
func (gp *Pin) Blink(value bool, blink time.Duration) {
	gp.c <- pinState{value: value, blink: blink}
}

// Set to value, and cycle between !value and value
func (gp *Pin) Cycle(value bool, cycle time.Duration) {
	gp.c <- pinState{value: value, blink: cycle, cycle: cycle}
}

func (gp *Pin) BlinkCycle(value bool, blink time.Duration, cycle time.Duration) {
	gp.c <- pinState{value: value, blink: blink, cycle: cycle}
}

func (gp *Pin) Close(wg *sync.WaitGroup) {
	if wg != nil {
		wg.Add(1)

		gp.closeWg = wg
	}

	close(gp.c)
}
