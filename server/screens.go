package server

import (
    "github.com/qmsk/e2/client"
    "fmt"
)

type Screens struct {
    client      *client.Client

    screenMap   map[string]Screen
}

func (screens *Screens) update() error {
    apiScreens, err := screens.client.ListScreens()
    if err != nil {
        return err
    }

    screenMap := make(map[string]Screen)

    for _, apiScreen := range apiScreens {
        screen := Screen{
            ID:     apiScreen.ID,
            Name:   apiScreen.Name,
        }

        screenMap[screen.String()] = screen
    }

    screens.screenMap = screenMap

    return nil
}

func (screens *Screens) Get() (interface{}, error) {
    if err := screens.update(); err != nil {
        return nil, err
    } else {
        return screens.screenMap, nil
    }
}

func (screens *Screens) Index(name string) (apiResource, error) {
    if err := screens.update(); err != nil {
        panic(err)
    } else if screen, found := screens.screenMap[name]; !found {
        return nil, nil
    } else {
        screenState := ScreenState{Screen: screen}

        if err := screenState.update(screens.client); err != nil {
            return screenState, err
        }

        return screenState, nil
    }
}

type Screen struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
}

func (self Screen) String() string {
    return fmt.Sprintf("%d", self.ID)
}

type ScreenState struct {
    Screen

    Program     []string        `json:"program"`
    Preview     []string        `json:"preview`
}

func (screenState *ScreenState) update(client *client.Client) error {
    screenContent, err := client.ListContent(screenState.ID)
    if err != nil {
        return err
    }

    for _, layer := range screenContent.Layers {
        if layer.LastSrcIdx <0 {
            continue
        }

        if layer.PgmMode > 0 {
            screenState.Program = append(screenState.Program, fmt.Sprintf("%d", layer.LastSrcIdx))
        }
        if layer.PvwMode > 0 {
            screenState.Preview = append(screenState.Preview, fmt.Sprintf("%d", layer.LastSrcIdx))
        }
    }

    return nil
}

func (screenState ScreenState) Get() (interface{}, error) {
    return screenState, nil
}
