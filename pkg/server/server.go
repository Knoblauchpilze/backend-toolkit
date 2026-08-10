package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	om "github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine          *gin.Engine
	log             *slog.Logger
	basePath        string
	port            uint16
	shutdownTimeout time.Duration
	router          *gin.RouterGroup
	stopChan        chan struct{}
}

func NewWithLogger(config Config, log *slog.Logger) *Server {
	engine := createGinEngine(log)

	s := &Server{
		engine:          engine,
		log:             log,
		basePath:        config.BasePath,
		port:            config.Port,
		shutdownTimeout: config.ShutdownTimeout,
		router:          engine.Group(""),
		stopChan:        make(chan struct{}, 1),
	}

	return s
}

func (s *Server) AddRoute(route *rest.Route) error {
	path := rest.ConcatenateEndpoints(s.basePath, route.Path())
	middlewares := buildMiddlewaresForRoute(route)
	handlers := append(middlewares, route.Handler())

	switch route.Method() {
	case http.MethodGet:
		s.router.GET(path, handlers...)
	case http.MethodPost:
		s.router.POST(path, handlers...)
	case http.MethodDelete:
		s.router.DELETE(path, handlers...)
	case http.MethodPatch:
		s.router.PATCH(path, handlers...)
	default:
		return ErrUnsupportedMethod
	}

	s.log.Debug("Registered route", slog.String("method", route.Method()), slog.String("path", path))

	return nil
}

func (s *Server) Start() error {
	address := fmt.Sprintf(":%d", s.port)

	s.log.Info("Starting server", slog.String("address", address))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-s.stopChan
		cancel()
	}()

	srv := &http.Server{
		Addr:    address,
		Handler: s.engine,
	}
	shutdownErrChan := make(chan error, 1)

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.log.Error("Server shutdown failed", slog.String("address", address), slog.Any("error", err))
			shutdownErrChan <- err
			return
		}

		shutdownErrChan <- nil
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.log.Error("Server failed", slog.String("address", address), slog.Any("error", err))
		return err
	}

	shutdownErr := <-shutdownErrChan
	if shutdownErr != nil {
		return shutdownErr
	}

	s.log.Info("Server gracefully shutdown", slog.String("address", address))

	return nil
}

func (s *Server) Stop() error {
	s.stopChan <- struct{}{}
	return nil
}

func createGinEngine(log *slog.Logger) *gin.Engine {
	e := gin.New()

	e.Use(func(c *gin.Context) {
		ctx := rest.WithContextLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	registerBaseMiddlewares(e, log)

	return e
}

func registerBaseMiddlewares(e *gin.Engine, log *slog.Logger) {
	// https://stackoverflow.com/questions/74020538/cors-preflight-did-not-succeed
	// https://stackoverflow.com/questions/6660019/restful-api-methods-head-options
	corsConfig := cors.Config{
		// https://www.stackhawk.com/blog/golang-cors-guide-what-it-is-and-how-to-enable-it/
		AllowAllOrigins: true,
		AllowMethods: []string{
			http.MethodOptions,
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
		},
	}

	e.Use(cors.New(corsConfig))
	e.Use(om.RequestId())
	e.Use(om.RequestTracer(log))
	e.Use(om.RequestLogger())
	e.Use(om.ErrorConverter())
	e.Use(om.Recover())
}

func buildMiddlewaresForRoute(route *rest.Route) []gin.HandlerFunc {
	out := []gin.HandlerFunc{}

	if route.UseResponseEnvelope() {
		out = append(out, om.ResponseEnvelope())
	}

	return out
}
