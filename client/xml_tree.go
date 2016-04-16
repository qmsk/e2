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

// XML-based collection tree unmarshalling state
type xmlCol struct {
    colMap      interface{}     // *map[int]T

    mapValue    reflect.Value   // existing map to read items for update
    itemName    string          // XML element name for type
    itemType    reflect.Type    // type of new items

    newMap      reflect.Value   // new map to write items 

    // decoding scope
    scope    xmlColScope
}

// setup xmlCol from reflect on the given colMap
func makeXmlCol(colMap interface{}) (xmlCol xmlCol, err error) {
    xmlCol.colMap = colMap

    ptrValue := reflect.ValueOf(xmlCol.colMap)

    if ptrValue.Kind() != reflect.Ptr {
        return xmlCol, fmt.Errorf("xmlCol.colMap must be *map[int]...")
    }

    xmlCol.mapValue = ptrValue.Elem()
    mapType := xmlCol.mapValue.Type()

    if mapType.Kind() != reflect.Map || mapType.Key().Kind() != reflect.Int {
        return xmlCol, fmt.Errorf("xmlCol.colMap must be *map[int]...")
    }

    xmlCol.itemType = mapType.Elem()
    xmlCol.itemName = xmlCol.itemType.Name() // matching element name

    // prepare copy-on-write map for update
    xmlCol.newMap = reflect.MakeMap(mapType)

    if !xmlCol.mapValue.IsNil() {
        // copy
        for _, keyValue := range xmlCol.mapValue.MapKeys() {
            xmlCol.newMap.SetMapIndex(keyValue, xmlCol.mapValue.MapIndex(keyValue))
        }
    }

    return
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

// unmarshal a <Foo> element
func (xmlCol *xmlCol) unmarshalItem(d *xml.Decoder, e xml.StartElement) error {
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

// Unmarshal a single <Foo> or an <Add>/n<Remove> collection of elements
func (xmlCol *xmlCol) unmarshalElement(d *xml.Decoder, e xml.StartElement) error {
    switch e.Name.Local {
        case "Add", "Remove":
            if err := xmlCol.setScope(e.Name.Local); err != nil {
                return err
            }

            // recurse
            return xmlCol.unmarshalElements(d, e)

        case xmlCol.itemName:
            // single item within scope
            return xmlCol.unmarshalItem(d, e)

        default:
            return fmt.Errorf("Unexpected StartElement <%s>", e.Name.Local)
    }
}

// Unmarshal current StartElement containing sub-elements up to matching EndElement
func (xmlCol *xmlCol) unmarshalElements(d *xml.Decoder, e xml.StartElement) error {
    for {
        xmlToken, err := d.Token()
        if err != nil {
            return err
        }

        // log.Printf("unmarshalElements %v: %#v", e.Name, xmlToken)

        switch t := xmlToken.(type) {
        case xml.CharData:
            if len(bytes.TrimSpace(t)) > 0 {
                return fmt.Errorf("Unexpected <%s> CharData: %v", e.Name.Local, t)
            }

        case xml.StartElement:
            // recurse into element
            if err := xmlCol.unmarshalElement(d, t); err != nil {
                return err
            }

        case xml.EndElement:
            if xmlCol.scope != "" {
                // internal <Add/Remove> scoping state
                if string(xmlCol.scope) == t.Name.Local {
                    // exit out of scope
                    xmlCol.scope = ""
                } else {
                    return fmt.Errorf("Unexpected <%s> scope EndElement </%s>", e.Name.Local, t.Name.Local)
                }
            }

            // done
            return nil

        default:
            return fmt.Errorf("Unexpected token: %#v\n", xmlToken)
        }
    }
}

// commit changes to col
func (xmlCol *xmlCol) commit() error {
    // replace map
    xmlCol.mapValue.Set(xmlCol.newMap)

    return nil
}

// Unmarshal complete XML collection elements of the form:
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
func unmarshalXMLCol(colMap interface{}, d *xml.Decoder, e xml.StartElement) error {
    // e is the <FooCol>, which we ignore except for error messages
    if xmlCol, err := makeXmlCol(colMap); err != nil {
        return err
    } else if err := xmlCol.unmarshalElements(d, e); err != nil {
        return err
    } else {
        return xmlCol.commit()
    }
}

// Unmarshal a single XML collection element of one of the forms:
//      <Add>
//          <Foo id="...">
//      <Foo id="...">
//      <Remove>
//          <Foo id="...">
func unmarshalXMLItem(colMap interface{}, d *xml.Decoder, e xml.StartElement) error {
    if xmlCol, err := makeXmlCol(colMap); err != nil {
        return err
    } else if err := xmlCol.unmarshalElement(d, e); err != nil {
        return err
    } else {
        return xmlCol.commit()
    }
}


