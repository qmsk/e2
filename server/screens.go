package server

import (
    "github.com/qmsk/e2/client"
    "fmt"
    "sort"
)

type Screens struct {
    client      *client.Client

    screenMap   map[string]Screen
}

func (screens *Screens) load(client *client.Client) error {
    apiScreens, err := client.ListScreenDestinations()
    if err != nil {
        return err
    }

    screenMap := make(map[string]Screen)

    for _, apiScreen := range apiScreens {
        screen := Screen{
            ID:     apiScreen.ID,
            Name:   apiScreen.Name,

            Dimensions: Dimensions{
                Width:      apiScreen.HSize,
                Height:     apiScreen.VSize,
            },
        }

        screenMap[screen.String()] = screen
    }

    screens.screenMap = screenMap

    return nil
}

func (screens *Screens) Get() (interface{}, error) {
    return screens.screenMap, nil
}

func (screens *Screens) Index(name string) (apiResource, error) {
    if screen, found := screens.screenMap[name]; !found {
        return nil, nil
    } else if screenState, err := screen.loadState(screens.client); err != nil {
        return screenState, err
    } else {
        return screenState, nil
    }
}

type Screen struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
    Dimensions  Dimensions  `json:"dimensions"`
}

func (screen Screen) loadState(client *client.Client) (ScreenState, error) {
    screenState := ScreenState{Screen: screen}

    return screenState, screenState.load(client)
}

func (self Screen) String() string {
    return fmt.Sprintf("%d", self.ID)
}

type ScreenState struct {
    Screen

    ProgramSources  []string        `json:"program_sources"`
    PreviewSources  []string        `json:"preview_sources"`
}

func (screenState *ScreenState) load(client *client.Client) error {
    screenContent, err := client.ListContent(screenState.ID)
    if err != nil {
        return err
    }

    for _, layer := range screenContent.Layers {
        if layer.LastSrcIdx <0 {
            continue
        }

        if layer.PgmMode > 0 {
            screenState.ProgramSources = append(screenState.ProgramSources, fmt.Sprintf("%d", layer.LastSrcIdx))
        }
        if layer.PvwMode > 0 {
            screenState.PreviewSources = append(screenState.PreviewSources, fmt.Sprintf("%d", layer.LastSrcIdx))
        }
    }

    sort.Strings(screenState.ProgramSources)
    sort.Strings(screenState.PreviewSources)

    return nil
}

func (screenState ScreenState) Get() (interface{}, error) {
    return screenState, nil
}
