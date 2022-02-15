package cache

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config interface {
	GetCacheCapacity() int64
	GetCachePath() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type lruCache struct {
	capacity int64
	queue    List
	items    map[string]*ListItem
	path     string
	logger   Logger
	mutex    sync.RWMutex
}

type cacheItem struct {
	key   string
	value string
}

var (
	ErrFileNameDecode = errors.New("unable to decode filename")
	ErrFileWrite      = errors.New("unable to write file to filesystem")
	ErrFileRemove     = errors.New("unable to remove file from filesystem")
	ErrFileRead       = errors.New("unable to read file from filesystem")
	ErrItemNotExists  = errors.New("cache item does not exist")
)

// New is a cache constructor: returns lruCache instance pointer.
func New(config Config, logger Logger) (*lruCache, error) {
	cache := lruCache{
		capacity: config.GetCacheCapacity(),
		path:     config.GetCachePath(),
		queue:    NewList(),
		items:    make(map[string]*ListItem, config.GetCacheCapacity()),
		logger:   logger,
	}

	err := os.MkdirAll(cache.path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

// Get is a LruCache getter: returns value if exists, or error, if doesnt.
func (l *lruCache) Get(key string) ([]byte, error) {
	l.mutex.Lock()
	item, exists := l.items[key]
	l.mutex.Unlock()

	// If cache element doesn't exist, return empty slice
	if !exists {
		return []byte{}, ErrItemNotExists
	}

	// If cache element exists, move it to front
	l.mutex.Lock()
	l.queue.MoveToFront(item)
	l.mutex.Unlock()

	// To get actual value, interface{} needs to be casted to cacheItem
	cacheItemElement := item.Value.(cacheItem)
	filename := cacheItemElement.value

	// Reading from filesystem
	value, err := l.readFromFileSystem(filename)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrFileRead, err)
	}

	return value, nil
}

// Set is a LruCache setter: sets or updates value, depends on whether the value exists or not.
func (l *lruCache) Set(key string, imageBytes []byte) error {
	l.mutex.Lock()
	listItem, exists := l.items[key]
	l.mutex.Unlock()

	filename := encodeFileName(key)
	cacheItemElement := cacheItem{key, filename}

	if exists {
		// If cache element exists, move it to front
		listItem.Value = cacheItemElement
		l.queue.MoveToFront(listItem)
	} else {
		// If cache element doesn't exist, create
		l.mutex.Lock()
		listItem = l.queue.PushFront(cacheItemElement)
		l.mutex.Unlock()

		// If list exceeds capacity, remove last element from list and map
		if int64(l.queue.Len()) > l.capacity {
			l.removeLastRecentUsedElement()
		}

		// Saving file to filesystem
		err := l.saveToFileSystem(filename, imageBytes)
		if err != nil {
			l.logger.Error(fmt.Errorf("%w: %s", ErrFileWrite, err))
		}
	}

	// Update map value anyway
	l.mutex.Lock()
	l.items[key] = listItem
	l.mutex.Unlock()

	return nil
}

func (l *lruCache) removeLastRecentUsedElement() {
	item := l.queue.Back()

	if item != nil {
		backCacheItem := item.Value.(cacheItem)
		filename := backCacheItem.value

		l.mutex.Lock()
		delete(l.items, backCacheItem.key)
		l.mutex.Unlock()
		l.queue.Remove(item)

		// Removing expired file from filesystem.
		err := l.removeFromFileSystem(filename)
		if err != nil {
			l.logger.Error(fmt.Errorf("%w: %s", ErrFileRemove, err))
		}
	}
}

func encodeFileName(key string) string {
	return base64.StdEncoding.EncodeToString([]byte(key))
}

//nolint:unused,deadcode
func decodeFileName(input string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrFileNameDecode, err)
	}

	return string(bytes), nil
}

func (l *lruCache) saveToFileSystem(filename string, bytes []byte) error {

	absFilename := filepath.Join(l.path, filename)
	return ioutil.WriteFile(absFilename, bytes, 0o644)
}

func (l *lruCache) readFromFileSystem(filename string) ([]byte, error) {

	absFilename := filepath.Join(l.path, filename)
	bytes, err := ioutil.ReadFile(absFilename)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (l *lruCache) removeFromFileSystem(filename string) error {

	absFilename := filepath.Join(l.path, filename)
	err := os.Remove(absFilename)

	return err
}

// Clear re-init lruCache instance.
func (l *lruCache) Clear() {
	l.queue = NewList()
	l.items = make(map[string]*ListItem, l.capacity)
}
