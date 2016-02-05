package client

import (
    "log"
)

type cacheValue interface{}
type cacheMap map[int]cacheValue

func (cacheMap *cacheMap) add(class string, id int, value interface{}) {
    log.Printf("Add %s %d: %v\n", class, id, value)
}

func (cacheMap *cacheMap) update(class string, id int, value interface{}) {
    log.Printf("Update %s %d: %v\n", class, id, value)
}

func (cacheMap *cacheMap) remove(class string, id int, value interface{}) {
    log.Printf("Remove %s %d: %v\n", class, id, value)
}

func (cacheMap *cacheMap) apply(class string, updateMap cacheMap) {
    if *cacheMap != nil {
        for id, value := range updateMap {
            if prev, exists := (*cacheMap)[id]; !exists {
                cacheMap.add(class, id, value)
            } else if value != prev {
                cacheMap.update(class, id, value)
            }
        }

        for id, value := range *cacheMap {
            if _, exists := updateMap[id]; !exists {
                cacheMap.remove(class, id, value)
            }
        }
    }

    *cacheMap = updateMap
}
