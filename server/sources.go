package server

import (
    "github.com/qmsk/e2/client"
    "fmt"
)

type Sources struct {
    client      *client.Client

    sourceMap   map[string]Source
}

func (sources *Sources) load() error {
    clientSources, err := sources.client.ListSources()
    if err != nil {
        return err
    }

    sources.sourceMap = make(map[string]Source)

    for _, apiSource := range clientSources {
        source := Source{
            ID:     apiSource.ID,
            Name:   apiSource.Name,
            Type:   apiSource.Type.String(),

            Dimensions: Dimensions{
                Width:      apiSource.HSize,
                Height:     apiSource.VSize,
            },
        }

        if apiSource.InputCfgIndex >= 0 {
            source.Status = apiSource.InputVideoStatus.String()
        }

        sources.sourceMap[source.String()] = source
    }

    return nil
}

func (sources *Sources) Index(name string) (apiResource, error) {
    if err := sources.load(); err != nil  {
        panic(err)
    } else if source, found := sources.sourceMap[name]; !found {
        return nil, nil
    } else {
        sourceState := SourceState{Source: source}

        if err := sourceState.load(sources.client); err != nil {
            panic(err)
        }

        return sourceState, nil
    }
}

func (sources *Sources) Get() (interface{}, error) {
    if err := sources.load(); err != nil {
        return nil, err
    } else {
       return sources.sourceMap, nil
    }
}

type Source struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Status      string      `json:"status,omitempty"`
    Dimensions  Dimensions  `json:"dimensions"`
}

func (source Source) String() string {
    return fmt.Sprintf("%d", source.ID)
}

func (source Source) Get() (interface{}, error) {
    return source, nil
}

type SourceState struct {
    Source

    Program     []string        `json:"program,omitempty"`
    Preview     []string        `json:"preview,omitempty"`
}

// XXX: expensive
func (sourceState *SourceState) load(client *client.Client) error {
    listDestinations, err := client.ListDestinations()
    if err != nil {
        return err
    }

    for _, screenDest := range listDestinations.ScreenDestinations {
        screenContent, err := client.ListContent(screenDest.ID)
        if err != nil {
            return err
        }

        for _, layer := range screenContent.Layers {
            if layer.LastSrcIdx != sourceState.ID {
                continue
            }

            if layer.PgmMode > 0 {
                sourceState.Program = append(sourceState.Program, fmt.Sprintf("%d", screenDest.ID))
            }
            if layer.PvwMode > 0 {
                sourceState.Preview = append(sourceState.Preview, fmt.Sprintf("%d", screenDest.ID))
            }
        }
    }

    return nil
}

func (source SourceState) Get() (interface{}, error) {
    return source, nil
}
