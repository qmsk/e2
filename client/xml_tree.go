package client

import (
    "bytes"
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

// Support for 
//  <FooCol> 
//      <Add>
//          <Foo id="...">
//      <Foo id="...">
type xmlCol struct {
    colMap  interface{} // *map[int]T
}

func unmarshalXMLMap(colMap interface{}, d *xml.Decoder, e xml.StartElement) error {
    return xmlCol{colMap: colMap}.UnmarshalXML(d, e)
}

// unmarshal an <Foo> element
func (xmlCol xmlCol) unmarshalItem(d *xml.Decoder, e xml.StartElement) error {
    mapValue := reflect.ValueOf(xmlCol.colMap).Elem()
    mapType := mapValue.Type()

    itemType := mapType.Elem()

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

// Unmarshal the top-level <FooCol> element
func (xmlCol xmlCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    ptrValue := reflect.ValueOf(xmlCol.colMap)

    mapValue := ptrValue.Elem()
    mapType := mapValue.Type()

    if mapType.Kind() != reflect.Map || mapType.Key().Kind() != reflect.Int {
        panic(fmt.Errorf("xmlMap should be map[int]..."))
    }

    itemType := mapType.Elem()
    itemName := itemType.Name() // matching element name

    for {
        if xmlToken, err := d.Token(); err != nil {
            return err
        } else if charData, valid := xmlToken.(xml.CharData); valid {
            if len(bytes.TrimSpace(charData)) > 0 {
                return fmt.Errorf("Unexpected <%s> CharData: %v", e.Name.Local, charData)
            }
        } else if startElement, valid := xmlToken.(xml.StartElement); valid {
            switch startElement.Name.Local {
                case "Add":
                    return fmt.Errorf("TODO <Add>")

                case "Remove":
                    return fmt.Errorf("TODO <Remove>")

                case itemName:
                    if err := xmlCol.unmarshalItem(d, startElement); err != nil {
                        return err
                    }

                default:
                    return fmt.Errorf("Unexpected <%s> StartElement <%s>", e.Name.Local, startElement.Name.Local)
            }
        } else if _, valid := xmlToken.(xml.EndElement); valid {
            break
        } else {
            return fmt.Errorf("Unexpected token: %#v\n", xmlToken)
        }
    }

    return nil
}
