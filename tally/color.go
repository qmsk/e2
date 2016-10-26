package tally

import (
	"encoding/json"

	"github.com/lucasb-eyer/go-colorful"
)

type Color struct {
	colorful.Color
}

func (color Color) Blend(blend Color, factor float64) Color {
	return Color{Color: color.BlendHsv(blend.Color, factor)}
}

func (color *Color) UnmarshalFlag(value string) error {
	if c, err := colorful.Hex("#" + value); err != nil {
		return err
	} else {
		*color = Color{Color: c}
	}

	return nil
}

func (color Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(color.Color.Hex())
}
