package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	om "github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Server interface {
	AddRoute(route rest.Route) error
	Start() error
	Stop() error
}

type serverImpl struct {
	echo            *echo.Echo
	basePath        string
	port            uint16
	shutdownTimeout time.Duration
	router          *echo.Group
	stopChan        chan struct{}
}

func NewWithLogger(config Config, log *slog.Logger) Server {
	echoServer := createEchoServer(log)

	s := &serverImpl{
		echo:            echoServer,
		basePath:        config.BasePath,
		port:            config.Port,
		shutdownTimeout: config.ShutdownTimeout,
		router:          echoServer.Group(""),
		stopChan:        make(chan struct{}, 1),
	}

	return s
}

func (s *serverImpl) AddRoute(route rest.Route) error {
	path := rest.ConcatenateEndpoints(s.basePath, route.Path())
	middlewares := buildMiddlewaresForRoute(route)

	switch route.Method() {
	case http.MethodGet:
		s.router.GET(path, route.Handler(), middlewares...)
	case http.MethodPost:
		s.router.POST(path, route.Handler(), middlewares...)
	case http.MethodDelete:
		s.router.DELETE(path, route.Handler(), middlewares...)
	case http.MethodPatch:
		s.router.PATCH(path, route.Handler(), middlewares...)
	default:
		return errors.NewCode(UnsupportedMethod)
	}

	s.echo.Logger.Debug("Registered route", slog.String("method", route.Method()), slog.String("path", path))

	return nil
}

func (s *serverImpl) Start() error {
	address := fmt.Sprintf(":%d", s.port)

	s.echo.Logger.Info("Starting server", slog.String("address", address))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-s.stopChan
		cancel()
	}()

	sc := echo.StartConfig{
		Address:         address,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: s.shutdownTimeout,
	}

	if err := sc.Start(ctx, s.echo); err != nil {
		s.echo.Logger.Error("Server failed", slog.String("address", address), slog.Any("error", err))
		return err
	}

	s.echo.Logger.Info("Server gracefully shutdown", slog.String("address", address))

	return nil
}

func (s *serverImpl) Stop() error {
	s.stopChan <- struct{}{}
	return nil
}

func createEchoServer(log *slog.Logger) *echo.Echo {
	e := echo.New()
	e.Logger = log

	registerBaseMiddlewares(e)

	return e
}

func registerBaseMiddlewares(e *echo.Echo) {
	// https://stackoverflow.com/questions/74020538/cors-preflight-did-not-succeed
	// https://stackoverflow.com/questions/6660019/restful-api-methods-head-options
	corsConf := middleware.CORSConfig{
		// https://www.stackhawk.com/blog/golang-cors-guide-what-it-is-and-how-to-enable-it/
		// Same as the default value
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodOptions,
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
		},
	}

	e.Use(middleware.CORSWithConfig(corsConf))
	e.Use(om.RequestLogger())
}
