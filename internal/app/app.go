package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
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
	ErrDownload        = errors.New("unable to download a file")
	ErrResize          = errors.New("unable to resize a file")
	ErrServerNotExists = errors.New("remove server doesn't exist")
	ErrRequest         = errors.New("request error")
	ErrFileRead        = errors.New("unable to read a file")
)

// downloadByURL downloads image by given url forwarding original headers.
func (app *Application) downloadByURL(url string, header http.Header) ([]byte, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, DefaultScheme+url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrRequest, err)
	}

	// Forwarding original headers to remote server.
	for name, values := range header {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			return []byte{}, fmt.Errorf("%w: %s", ErrServerNotExists, err)
		}

		return []byte{}, fmt.Errorf("%w: %s", ErrDownload, err)
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrFileRead, err)
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
		return []byte{}, err
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
