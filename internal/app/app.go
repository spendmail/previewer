package app

import (
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

type Application struct {
	Logger  Logger
	Resizer Resizer
}

var (
	ErrDownload = errors.New("unable to download a file")
	ErrResize   = errors.New("unable to resize a file")
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

	sourceBytes, err := app.downloadByUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrDownload, err)
	}

	resultBytes, err := app.Resizer.Resize(uint(width), uint(height), sourceBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrResize, err)
	}

	return resultBytes, nil
}

func New(logger Logger, resizer Resizer) (*Application, error) {
	return &Application{
		Logger:  logger,
		Resizer: resizer,
	}, nil
}
