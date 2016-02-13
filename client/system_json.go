package client

import (
    "fmt"
    "encoding/json"
    "reflect"
)


// Marshal map[int]... to JSON {"id": { ... }}
func marshalJSONMap(colMap interface{}) ([]byte, error) {
    mapValue := reflect.ValueOf(colMap)
    mapType := mapValue.Type()

    if mapType.Kind() != reflect.Map || mapType.Key().Kind() != reflect.Int {
        panic(fmt.Errorf("colMap must be *map[int]..."))
    }

    jsonMap := make(map[string]interface{})

    for _, idValue := range mapValue.MapKeys() {
        itemValue := mapValue.MapIndex(idValue)

        jsonMap[fmt.Sprintf("%d", idValue.Int())] = itemValue.Interface()
    }

    return json.Marshal(jsonMap)
}
