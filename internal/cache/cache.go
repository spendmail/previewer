package cache

import "os"

type Config interface {
	GetCacheCapacity() int64
	GetCachePath() string
}

// Set is a LruCache setter: sets or updates value, depends on whether the value it exists or not.
func (l *lruCache) Set(key string, value []byte) bool {
	//listItem, exists := l.items[key]
	//cacheItemElement := cacheItem{string(key), value}
	//
	//if exists {
	//	// If cache element exists, move it to front
	//	listItem.Value = cacheItemElement
	//	l.queue.MoveToFront(listItem)
	//} else {
	//	// If cache element doesn't exist, create
	//	listItem = l.queue.PushFront(cacheItemElement)
	//
	//	// If list exceeds capacity, remove last element from list and map
	//	if int64(l.queue.Len()) > l.capacity {
	//		item := l.queue.Back()
	//		backCacheItem := item.Value.(cacheItem)
	//		delete(l.items, backCacheItem.key)
	//		l.queue.Remove(item)
	//	}
	//}
	//
	//// Update map value anyway
	//l.items[key] = listItem
	//
	//return exists
	return false
}

// Get is a LruCache getter: returns value if exists, or nil, if doesnt.
func (l *lruCache) Get(key string) ([]byte, bool) {
	//item, exists := l.items[key]
	//
	//// If cache element doesn't exist, return nil, false
	//if !exists {
	//	return nil, exists
	//}
	//
	//// If cache element exists, moves it to front
	//l.queue.MoveToFront(item)
	//
	//// To get actual value, interface{} needs to be casted to cacheItem
	//cacheItemElement := item.Value.(cacheItem)
	//return cacheItemElement.value, exists
	return nil, false
}

// Clear reinit lruCache instance.
func (l *lruCache) Clear() {
	//l.queue = NewList()
	//l.items = make(map[string]*ListItem, l.capacity)
}

type lruCache struct {
	capacity int64
	queue    List
	items    map[string]*ListItem
	path     string
}

type cacheItem struct {
	key   string
	value []byte
}

// New is a cache constructor: returns lruCache instance pointer.
func New(config Config) (*lruCache, error) {
	cache := lruCache{
		capacity: config.GetCacheCapacity(),
		path:     config.GetCachePath(),
		queue:    NewList(),
		items:    make(map[string]*ListItem, config.GetCacheCapacity()),
	}

	err := os.MkdirAll(cache.path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}
