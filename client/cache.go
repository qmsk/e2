package client

import (
    "fmt"
    "log"
    "reflect"
)

type cacheObject interface{
    cacheID()   int
}

type cacheMap map[int]cacheObject

func (cacheMap *cacheMap) add(object cacheObject) {
    log.Printf("Add %T %d: %v\n", object, object.cacheID(), object)
}

func (cacheMap *cacheMap) update(object cacheObject) {
    log.Printf("Update %T %d: %v\n", object, object.cacheID(), object)
}

func (cacheMap *cacheMap) remove(object cacheObject) {
    log.Printf("Remove %T %d: %v\n", object, object.cacheID(), object)
}

func (cacheMap *cacheMap) applyMap(updateMap cacheMap) {
    if *cacheMap != nil {
        for id, object := range updateMap {
            if prev, exists := (*cacheMap)[id]; !exists {
                cacheMap.add(object)
            } else if object != prev {
                cacheMap.update(object)
            }
        }

        for id, object := range *cacheMap {
            if _, exists := updateMap[id]; !exists {
                cacheMap.remove(object)
            }
        }
    }

    *cacheMap = updateMap
}

// Update cache set from `[]cacheObject`
func (self *cacheMap) apply(set interface{}) {
    setValue := reflect.ValueOf(set)

    switch setValue.Kind() {
    case reflect.Slice:
        updateMap := make(cacheMap)

        for i := 0; i < setValue.Len(); i++ {
            cacheObject := setValue.Index(i).Interface().(cacheObject)

            updateMap[cacheObject.cacheID()] = cacheObject
        }

        self.applyMap(updateMap)

    default:
        panic(fmt.Errorf("Invalid value %T: %#v", set, set))
    }
}
