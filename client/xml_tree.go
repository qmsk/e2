package client

import (
    "fmt"
    "reflect"
    "encoding/xml"
)

func xmlAttr(e xml.StartElement, name string) (value string) {
    for _, attr := range e.Attr {
        if attr.Name.Local == name {
            return attr.Value
        }
    }

    return ""
}

func xmlID(e xml.StartElement) (id int, err error) {
    value := xmlAttr(e, "id")

    if _, err := fmt.Sscanf(value, "%d", &id); err != nil {
        return id, err
    } else {
        return id, nil
    }
}

type xmlMap interface{} // *map[int]T

func unmarshalXMLMap(x xmlMap, d *xml.Decoder, e xml.StartElement) error {
    ptrValue := reflect.ValueOf(x)

    mapValue := ptrValue.Elem()
    mapType := mapValue.Type()

    if mapType.Kind() != reflect.Map || mapType.Key().Kind() != reflect.Int {
        panic(fmt.Errorf("xmlMap should be map[int]..."))
    }

    itemType := mapType.Elem()

    if e.Name.Local != itemType.Name() {
        return fmt.Errorf("xmpMap Element <%s> mismatch; should be type %s", e.Name.Local, itemType.Name())
    }

    // index by id
    id, err := xmlID(e)
    if err != nil {
        return err
    }

    idValue := reflect.ValueOf(id)

    // update into new map
    newMap := reflect.MakeMap(mapType)

    if !mapValue.IsNil() {
        // copy
        for _, keyValue := range mapValue.MapKeys() {
            newMap.SetMapIndex(keyValue, mapValue.MapIndex(keyValue))
        }
    }
    mapValue.Set(newMap)

    // unmarshal into existing item from map, or zero value if item was not in map
    itemValue := reflect.New(itemType)

    if getValue := mapValue.MapIndex(idValue); getValue.IsValid() {
        itemValue.Elem().Set(getValue)
    }

    if err := d.DecodeElement(itemValue.Interface(), &e); err != nil {
        return err
    }

    // store into map
    mapValue.SetMapIndex(idValue, itemValue.Elem())

    return nil
}
