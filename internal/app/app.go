package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultScheme = "http://"
)

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
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Clear()
}

type Application struct {
	Logger  Logger
	Resizer Resizer
	Cache   Cache
}

var (
	ErrDownload = errors.New("unable to download a file")
	ErrResize   = errors.New("unable to resize a file")
)

// downloadByURL downloads image by given url forwarding original headers.
func (app *Application) downloadByURL(url string, header http.Header) ([]byte, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, DefaultScheme+url, nil)
	if err != nil {
		return []byte{}, err
	}

	// Forwarding original headers to remote server.
	for name, values := range header {
		for _, value := range values {
			app.Logger.Error(fmt.Sprintf("%v: %v", name, value))
			request.Header.Add(name, value)
		}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if err != nil {
			return []byte{}, err
		}
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (app *Application) ResizeImageByURL(width, height int, url string, header http.Header) ([]byte, error) {
	// Key includes sizes in order to store different files for different sizes of the same file.
	cacheKey := fmt.Sprintf("%s-%d-%d", url, width, height)

	// If file exists in cache, return from there.
	resultBytes, err := app.Cache.Get(cacheKey)
	if err == nil {
		return resultBytes, nil
	}

	// Otherwise, download file.
	sourceBytes, err := app.downloadByURL(url, header)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrDownload, err)
	}

	// Process file.
	resultBytes, err = app.Resizer.Resize(uint(width), uint(height), sourceBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrResize, err)
	}

	// Set processed image in cache
	_ = app.Cache.Set(cacheKey, resultBytes)

	// And return slice of bytes.
	return resultBytes, nil
}

func New(logger Logger, resizer Resizer, cache Cache) (*Application, error) {
	return &Application{
		Cache:   cache,
		Logger:  logger,
		Resizer: resizer,
	}, nil
}
