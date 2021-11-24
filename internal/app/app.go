package app

import (
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"io"
	"net/http"
)

const (
	DefaultScheme = "http://"
)

type Config interface {
	GetCacheSize() int
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Resizer interface {
	Resize(width, height uint, image []byte) ([]byte, error)
}

type Cache interface {
	Set(key, value interface{}) error
	Get(key interface{}) (interface{}, error)
}

type Application struct {
	Logger  Logger
	Resizer Resizer
	Cache   Cache
}

var (
	ErrDownload = errors.New("unable to download a file")
	ErrResize   = errors.New("unable to resize a file")
	ErrCacheSet = errors.New("unable to set a cache value")
)

func (app *Application) downloadByUrl(url string) ([]byte, error) {

	resp, err := http.Get(DefaultScheme + url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (app *Application) ResizeImageByUrl(width, height int, url string) ([]byte, error) {

	i, err := app.Cache.Get(url)
	if err == nil {
		return i.([]byte), nil
	}

	sourceBytes, err := app.downloadByUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrDownload, err)
	}

	resultBytes, err := app.Resizer.Resize(uint(width), uint(height), sourceBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrResize, err)
	}

	err = app.Cache.Set(url, resultBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrCacheSet, err)
	}

	return resultBytes, nil
}

func New(config Config, logger Logger, resizer Resizer) (*Application, error) {
	return &Application{
		Cache:   gcache.New(config.GetCacheSize()).LRU().Build(),
		Logger:  logger,
		Resizer: resizer,
	}, nil
}
