package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	ginCors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	om "github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
)

type Server interface {
	AddRoute(route rest.Route) error
	Start() error
	Stop() error
}

type serverImpl struct {
	engine          *gin.Engine
	log             *slog.Logger
	basePath        string
	port            uint16
	shutdownTimeout time.Duration
	stopChan        chan struct{}
}

func NewWithLogger(config Config, log *slog.Logger) Server {
	engine := createGinEngine(log)

	return &serverImpl{
		engine:          engine,
		log:             log,
		basePath:        config.BasePath,
		port:            config.Port,
		shutdownTimeout: config.ShutdownTimeout,
		stopChan:        make(chan struct{}, 1),
	}
}

func (s *serverImpl) AddRoute(route rest.Route) error {
	path := rest.ConcatenateEndpoints(s.basePath, route.Path())
	middlewares := buildMiddlewaresForRoute(route)
	handlers := append(middlewares, wrapHandler(route.Handler()))

	switch route.Method() {
	case http.MethodGet:
		s.engine.GET(path, handlers...)
	case http.MethodPost:
		s.engine.POST(path, handlers...)
	case http.MethodDelete:
		s.engine.DELETE(path, handlers...)
	case http.MethodPatch:
		s.engine.PATCH(path, handlers...)
	default:
		return ErrUnsupportedMethod
	}

	s.log.Debug("Registered route", slog.String("method", route.Method()), slog.String("path", path))

	return nil
}

func (s *serverImpl) Start() error {
	address := fmt.Sprintf(":%d", s.port)

	s.log.Info("Starting server", slog.String("address", address))

	httpServer := &http.Server{
		Addr:    address,
		Handler: s.engine,
	}

	go func() {
		<-s.stopChan
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.log.Error("Server shutdown error", slog.String("address", address), slog.Any("error", err))
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Error("Server failed", slog.String("address", address), slog.Any("error", err))
		return err
	}

	s.log.Info("Server gracefully shutdown", slog.String("address", address))

	return nil
}

func (s *serverImpl) Stop() error {
	s.stopChan <- struct{}{}
	return nil
}

// wrapHandler adapts a rest.HandlerFunc (which returns an error) into a
// gin.HandlerFunc by storing any returned error in the Gin context so that
// the ErrorConverter middleware can convert it to an HTTP error response.
func wrapHandler(h rest.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			_ = c.Error(err)
		}
	}
}

func createGinEngine(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	registerBaseMiddlewares(engine, log)

	return engine
}

func registerBaseMiddlewares(engine *gin.Engine, log *slog.Logger) {
	corsConfig := ginCors.Config{
		AllowAllOrigins: true,
		AllowMethods: []string{
			http.MethodOptions,
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
		},
	}

	// Seed the Gin context with the server-level logger before other middlewares run.
	engine.Use(func(c *gin.Context) {
		om.SetContextLogger(c, log)
		c.Next()
	})
	engine.Use(ginCors.New(corsConfig))
	engine.Use(om.RequestLogger())
}
