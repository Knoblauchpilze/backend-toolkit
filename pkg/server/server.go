package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	om "github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
}

func NewWithLogger(config Config, log logger.Logger) Server {
	echoServer := createEchoServer(logger.Wrap(log))

	s := &serverImpl{
		echo:            echoServer,
		basePath:        config.BasePath,
		port:            config.Port,
		shutdownTimeout: config.ShutdownTimeout,
		router:          echoServer.Group(""),
	}

	return s
}

func (s *serverImpl) AddRoute(route rest.Route) error {
	path := rest.ConcatenateEndpoints(s.basePath, route.Path())

	switch route.Method() {
	case http.MethodGet:
		s.router.GET(path, route.Handler())
	case http.MethodPost:
		s.router.POST(path, route.Handler())
	case http.MethodDelete:
		s.router.DELETE(path, route.Handler())
	case http.MethodPatch:
		s.router.PATCH(path, route.Handler())
	default:
		return errors.NewCode(UnsupportedMethod)
	}

	s.echo.Logger.Debugf("Registered %s %s", route.Method(), path)

	return nil
}

func (s *serverImpl) Start() error {
	// https://echo.labstack.com/docs/cookbook/graceful-shutdown
	address := fmt.Sprintf(":%d", s.port)

	s.echo.Logger.Infof("Starting server at %s", address)
	err := s.echo.Start(address)

	if err == http.ErrServerClosed {
		s.echo.Logger.Infof("Server at %s gracefully shutdown", address)
		return nil
	}

	s.echo.Logger.Infof("Server at %s failed with error: %v", address, err)

	return err
}

func (s *serverImpl) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	return s.echo.Shutdown(ctx)
}

func createEchoServer(log echo.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = log

	registerMiddlewares(e)

	return e
}

func registerMiddlewares(e *echo.Echo) {
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
	e.Use(middleware.Gzip())
	e.Use(om.RequestLogger())
	e.Use(om.ResponseEnvelope())
	e.Use(om.RequestTracer(e.Logger))
	e.Use(om.ErrorConverter())
	e.Use(om.Recover())
}
