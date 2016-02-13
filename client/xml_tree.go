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

type xmlColScope string

const (
    xmlColAdd       xmlColScope          = "Add"
    xmlColRemove                        = "Remove"
)

// Support for 
//  <FooCol> 
//      <Add>
//          <Foo id="...">
//      <Foo id="...">
type xmlCol struct {
    colMap  interface{} // *map[int]T

    // decoding scope
    scope    xmlColScope
}

func unmarshalXMLMap(colMap interface{}, d *xml.Decoder, e xml.StartElement) error {
    return xmlCol{colMap: colMap}.UnmarshalXML(d, e)
}

func (xmlCol *xmlCol) setScope(scope string) error {
    if xmlCol.scope != "" {
        return fmt.Errorf("Unexpected nested <%s>", scope)
    }

    switch xmlColScope(scope) {
    case xmlColAdd:
        xmlCol.scope = xmlColAdd

    case xmlColRemove:
        xmlCol.scope = xmlColRemove

    default:
        return fmt.Errorf("Invalid scope <%s>", scope)
    }

    return nil
}

// unmarshal an <Foo> element
func (xmlCol *xmlCol) unmarshalItem(d *xml.Decoder, e xml.StartElement) error {
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

    if xmlCol.scope == xmlColRemove {
        if err := d.Skip(); err != nil {
            return err
        }

        // delete identified element fom map
        newMap.SetMapIndex(idValue, reflect.Value{})
    } else {
        // unmarshal into existing item from map, or zero value if item was not in map
        itemValue := reflect.New(itemType)

        if xmlCol.scope == xmlColAdd {
            // there should never be any existing item
        } else if getValue := mapValue.MapIndex(idValue); getValue.IsValid() {
            itemValue.Elem().Set(getValue)
        }

        if err := d.DecodeElement(itemValue.Interface(), &e); err != nil {
            return err
        }

        // store into map
        newMap.SetMapIndex(idValue, itemValue.Elem())
    }

    // replace map
    mapValue.Set(newMap)

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
                case "Add", "Remove":
                    if err := xmlCol.setScope(startElement.Name.Local); err != nil {
                        return err
                    }

                case itemName:
                    // within scope
                    if err := xmlCol.unmarshalItem(d, startElement); err != nil {
                        return err
                    }

                default:
                    return fmt.Errorf("Unexpected <%s> StartElement <%s>", e.Name.Local, startElement.Name.Local)
            }
        } else if endElement, valid := xmlToken.(xml.EndElement); valid {
            if xmlCol.scope == "" {
                break
            } else if string(xmlCol.scope) == endElement.Name.Local {
                // exit out of <Add/Remove> scope
                xmlCol.scope = ""
            } else {
                return fmt.Errorf("Unexpected <%s> EndElement </%s>", e.Name.Local, endElement.Name.Local)
            }
        } else {
            return fmt.Errorf("Unexpected token: %#v\n", xmlToken)
        }
    }

    return nil
}
