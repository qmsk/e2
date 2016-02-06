package server

import (
    "github.com/qmsk/e2/client"
    "fmt"
)

type Presets struct {
    presetMap   map[string]Preset
}

func (presets *Presets) load(client *client.Client) error {
    apiPresets, err := client.ListPresets()
    if err != nil {
        return err
    }

    presetMap := make(map[string]Preset)

    for _, apiPreset := range apiPresets {
        // parse sno
        preset := Preset{
            ID:     apiPreset.ID,
            Name:   apiPreset.Name,

            Locked: apiPreset.LockMode > 0,
        }

        preset.Group, preset.Index = apiPreset.ParseOrder()

        presetMap[preset.String()] = preset
    }

    presets.presetMap = presetMap

    return nil
}

func (presets *Presets) Get() (interface{}, error) {
    return presets.presetMap, nil
}

func (presets *Presets) Index(name string) (apiResource, error) {
    if preset, found := presets.presetMap[name]; !found {
        return nil, nil
    } else {
        return preset, nil
    }
}

type Preset struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`

    // Decomposed Sequence order
    Group       int         `json:"group"`
    Index       int         `json:"index"`

    Locked      bool        `json:"locked"`
}

func (preset Preset) String() string {
    return fmt.Sprintf("%d", preset.ID)
}

func (preset Preset) Get() (interface{}, error) {
    return preset, nil
}
