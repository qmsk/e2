package client

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Marshal map[int]... to JSON {"id": { ... }}
func marshalJSONMap(colMap interface{}) ([]byte, error) {
	mapValue := reflect.ValueOf(colMap)
	mapType := mapValue.Type()

	if mapType.Kind() != reflect.Map {
		panic(fmt.Errorf("colMap must be *map[...]..."))
	}

	jsonMap := make(map[string]interface{})

	for _, idValue := range mapValue.MapKeys() {
		itemValue := mapValue.MapIndex(idValue)

		jsonMap[fmt.Sprintf("%v", idValue.Interface())] = itemValue.Interface()
	}

	return json.Marshal(jsonMap)
}

func marshalJSONString(value interface{}) ([]byte, error) {
	valueString := fmt.Sprintf("%v", value)

	return json.Marshal(valueString)
}

func marshalJSONInteger(value interface{}) ([]byte, error) {
	valueString := fmt.Sprintf("%d", value)

	return json.Marshal(valueString)
}

func marshalJSONNull() ([]byte, error) {
	return json.Marshal(nil)
}
