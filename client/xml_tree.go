package client

import (
    "bytes"
    "fmt"
    "log"
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

// Unmarshal XML messages of the form:
//  <FooCol> 
//      <Add>
//          <Foo id="...">
//      <Foo id="...">
//      <Remove>
//          <Foo id="...">
//
// The xmlCol should be created with a colMap pointing to a map. The existing map values will be copied, or a new map created if the map is still nil.
// *colMap will then be updated with the resulting new map.
// This copy-on-write mechanism ensures that the datastructure is safe for concurrent read access when a copy of it is sent via a chan.
type xmlCol struct {
    colMap      interface{} // *map[int]T

    mapValue    reflect.Value   // existing map to read items for update
    itemType    reflect.Type    // type of new items

    newMap      reflect.Value   // new map to write items 

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
func (xmlCol xmlCol) unmarshalItem(d *xml.Decoder, e xml.StartElement) error {
    // index by id
    id, err := xmlID(e)
    if err != nil {
        return err
    }

    idValue := reflect.ValueOf(id)

    if xmlCol.scope == xmlColRemove {
        log.Printf("XML remove %s[%d]\n", xmlCol.itemType.Name(), idValue.Interface())

        if err := d.Skip(); err != nil {
            return err
        }

        // delete identified element fom map
        xmlCol.newMap.SetMapIndex(idValue, reflect.Value{})
    } else {
        // unmarshal into existing item from map, or zero value if item was not in map
        itemValue := reflect.New(xmlCol.itemType)

        if xmlCol.scope == xmlColAdd {
            log.Printf("XML add %s[%d]\n", xmlCol.itemType.Name(), idValue.Interface())
            // there should never be any existing item
        } else if getValue := xmlCol.newMap.MapIndex(idValue); getValue.IsValid() {
            log.Printf("XML set %s[%d]\n", getValue.Type().Name(), idValue.Interface())

            itemValue.Elem().Set(getValue)
        } else {
            log.Printf("XML new %s[%d]\n", xmlCol.itemType.Name(), idValue.Interface())
        }

        if err := d.DecodeElement(itemValue.Interface(), &e); err != nil {
            return err
        }

        // store into map
        xmlCol.newMap.SetMapIndex(idValue, itemValue.Elem())
    }

    return nil
}

// Unmarshal the top-level <FooCol> element
func (xmlCol xmlCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    ptrValue := reflect.ValueOf(xmlCol.colMap)

    if ptrValue.Kind() != reflect.Ptr {
        panic(fmt.Errorf("xmlCol.colMap must be *map[int]..."))
    }

    xmlCol.mapValue = ptrValue.Elem()
    mapType := xmlCol.mapValue.Type()

    if mapType.Kind() != reflect.Map || mapType.Key().Kind() != reflect.Int {
        panic(fmt.Errorf("xmlCol.colMap must be *map[int]..."))
    }

    xmlCol.itemType = mapType.Elem()
    itemName := xmlCol.itemType.Name() // matching element name

    // prepare copy-on-write map for update
    xmlCol.newMap = reflect.MakeMap(mapType)

    if !xmlCol.mapValue.IsNil() {
        // copy
        for _, keyValue := range xmlCol.mapValue.MapKeys() {
            xmlCol.newMap.SetMapIndex(keyValue, xmlCol.mapValue.MapIndex(keyValue))
        }
    }

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

    // replace map
    xmlCol.mapValue.Set(xmlCol.newMap)

    return nil
}
