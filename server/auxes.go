package server

import (
    "github.com/qmsk/e2/client"
    "fmt"
)

type Auxes struct {
    auxMap   map[string]Aux
}

func (auxes *Auxes) load(client *client.Client) error {
    apiAuxes, err := client.ListAuxDestinations()
    if err != nil {
        return err
    }

    auxMap := make(map[string]Aux)

    for _, apiAux := range apiAuxes {
        aux := Aux{
            ID:     apiAux.ID,
            Name:   apiAux.Name,
        }

        auxMap[aux.String()] = aux
    }

    auxes.auxMap = auxMap

    return nil
}

func (auxes *Auxes) Get() (interface{}, error) {
    return auxes.auxMap, nil
}

func (auxes *Auxes) Index(name string) (apiResource, error) {
    if aux, found := auxes.auxMap[name]; !found {
        return nil, nil
    } else {
        return aux, nil
    }
}

type Aux struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
}

func (aux Aux) String() string {
    return fmt.Sprintf("%d", aux.ID)
}

func (aux Aux) Get() (interface{}, error) {
    return aux, nil
}
