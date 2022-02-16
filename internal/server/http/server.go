package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	internalapp "github.com/spendmail/previewer/internal/app"
)

const URLResizePattern = "/fill/{width:[0-9]+}/{height:[0-9]+}/{url:.+}"

type Config interface {
	GetHTTPHost() string
	GetHTTPPort() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Application interface {
	ResizeImageByURL(width, height int, url string, header http.Header) ([]byte, error)
}

type Server struct {
	Logger Logger
	Server *http.Server
}

var (
	ErrParameterParseWidth  = errors.New("unable to parse image width")
	ErrParameterParseHeight = errors.New("unable to parse image height")
	ErrResizeImage          = errors.New("unable to resize an image")
	ErrResponseWrite        = errors.New("unable to write a response")
)

type Handler struct {
	App    Application
	Logger Logger
}

func SendBadGatewayStatus(w http.ResponseWriter, h *Handler, errType, err error) {
	w.WriteHeader(http.StatusBadGateway)
	if n, e := w.Write([]byte(errType.Error())); e != nil {
		h.Logger.Error(fmt.Errorf("%w: trying to write %d bytes: %s", ErrResponseWrite, n, e.Error()))
	}
	h.Logger.Error(fmt.Errorf("%w: %s", errType, err.Error()))
}

func (h *Handler) resizeHandler(w http.ResponseWriter, r *http.Request) {
	width, err := strconv.Atoi(mux.Vars(r)["width"])
	if err != nil {
		SendBadGatewayStatus(w, h, ErrParameterParseWidth, err)
		return
	}

	height, err := strconv.Atoi(mux.Vars(r)["height"])
	if err != nil {
		SendBadGatewayStatus(w, h, ErrParameterParseHeight, err)
		return
	}

	bytes, err := h.App.ResizeImageByURL(width, height, mux.Vars(r)["url"], r.Header)
	if err != nil {
		switch {
		case errors.Is(err, internalapp.ErrDownload):
			SendBadGatewayStatus(w, h, internalapp.ErrDownload, err)
		case errors.Is(err, internalapp.ErrResize):
			SendBadGatewayStatus(w, h, internalapp.ErrResize, err)
		case errors.Is(err, internalapp.ErrServerNotExists):
			SendBadGatewayStatus(w, h, internalapp.ErrServerNotExists, err)
		case errors.Is(err, internalapp.ErrRequest):
			SendBadGatewayStatus(w, h, internalapp.ErrRequest, err)
		case errors.Is(err, internalapp.ErrFileRead):
			SendBadGatewayStatus(w, h, internalapp.ErrFileRead, err)
		default:
			SendBadGatewayStatus(w, h, ErrResizeImage, err)
		}

		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(bytes))
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	if _, err := w.Write(bytes); err != nil {
		h.Logger.Error(fmt.Errorf("%w: %s", ErrResizeImage, err.Error()))
	}
}

func New(config Config, logger Logger, app Application) *Server {
	handler := &Handler{
		App:    app,
		Logger: logger,
	}

	router := mux.NewRouter()
	router.HandleFunc(URLResizePattern, handler.resizeHandler).Methods("GET")

	server := &http.Server{
		Addr:    net.JoinHostPort(config.GetHTTPHost(), config.GetHTTPPort()),
		Handler: router,
	}

	return &Server{
		Logger: logger,
		Server: server,
	}
}

// Start launches a HTTP server.
func (s *Server) Start() error {
	return s.Server.ListenAndServe()
}

// Stop suspends HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
