package app

import (
	"sync"
	"time"
	"udpforward/logger"
)

type cacheItem struct {
	value     interface{}
	validTill time.Time
}

type Cache struct {
	cacheMap  *sync.Map
	cacheTime time.Duration
}

func cleanupCache(cache *Cache) {
	for {
		time.Sleep(time.Second * 60)
		now := time.Now()
		deleteNum := 0
		cache.cacheMap.Range(func(k interface{}, v interface{}) bool {
			vv := v.(*cacheItem)
			if now.After(vv.validTill) {
				cache.cacheMap.Delete(k)
				deleteNum++
			}
			return true
		})
		if deleteNum > 0 {
			logger.Debug("cleanup cache, delete %d items", deleteNum)
		}
	}
}

func NewCache(cacheSeconds int) *Cache {
	cache := &Cache{
		cacheMap:  &sync.Map{},
		cacheTime: time.Duration(cacheSeconds) * time.Second,
	}
	go cleanupCache(cache)
	return cache
}

func (c *Cache) Add(key string, value interface{}) {
	var item *cacheItem
	v, ok := c.cacheMap.Load(key)
	if ok {
		item = v.(*cacheItem)
		item.validTill = time.Now().Add(c.cacheTime)
	} else {
		item = &cacheItem{
			value:     value,
			validTill: time.Now().Add(c.cacheTime),
		}
		c.cacheMap.Store(key, item)
	}
}

func (c *Cache) Get(key string) interface{} {
	v, ok := c.cacheMap.Load(key)
	if !ok {
		return nil
	}
	item := v.(*cacheItem)
	item.validTill = time.Now().Add(c.cacheTime)
	return item.value
}
