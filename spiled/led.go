package spiled

import (
	"fmt"
	"log"
	"time"
)

// 24-bit colors + 8-bit intensity
type LED struct {
	Intensity	uint8
	Red			uint8
	Green		uint8
	Blue		uint8

	// animations
	strobeStart	 time.Time
	strobePeriod time.Duration
}

// Use #RRGGBBAA format
func (led LED) String() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", led.Red, led.Green, led.Blue, led.Intensity)
}

func (led *LED) UnmarshalFlag (value string) error {
	if _, err := fmt.Sscanf(value, "%02x%02x%02x%02x", &led.Red, &led.Green, &led.Blue, &led.Intensity); err == nil {

	} else if _, err := fmt.Sscanf(value, "%02x%02x%02x", &led.Red, &led.Green, &led.Blue); err == nil {
		led.Intensity = 0xff
	} else {
		return fmt.Errorf("Invalid LED RRGGBB[AA] color: %v", value)
	}

	return nil
}

func (led *LED) Strobe(period time.Duration) {
	led.strobeStart = time.Now()
	led.strobePeriod = period
}

func (led LED) render(renderTime time.Time) []byte {
	var intensity = led.Intensity

	if !led.strobeStart.IsZero() && !renderTime.IsZero() {
		strobeOffset := (renderTime.Sub(led.strobeStart) % led.strobePeriod).Seconds() / led.strobePeriod.Seconds()

		log.Printf("LED Strobe offset=%v", strobeOffset)

		intensity = uint8(strobeOffset * float64(intensity))
	}

	return []byte{
		0xC0 | (intensity >> 2), // convert to 6-bit
		led.Blue,
		led.Green,
		led.Red,
	}
}
