package spiled

import (
	"fmt"
)

// 24-bit colors + 8-bit intensity
type LED struct {
	Intensity	uint8
	Red			uint8
	Green		uint8
	Blue		uint8
}

// Use #RRGGBBAA format
func (led LED) String() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", led.Red, led.Green, led.Blue, led.Intensity)
}

func (led LED) Bytes() []byte {
	return []byte{
		0xC0 | (led.Intensity >> 2), // convert to 6-bit
		led.Blue,
		led.Green,
		led.Red,
	}
}
