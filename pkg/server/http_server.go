package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	engine          *gin.Engine
	log             *slog.Logger
	basePath        string
	shutdownTimeout time.Duration
	router          *gin.RouterGroup
}

func NewWithLogger(config Config, log *slog.Logger) *HttpServer {
	engine := createGinEngine(log)

	s := &HttpServer{
		engine:          engine,
		log:             log,
		basePath:        config.BasePath,
		shutdownTimeout: config.ShutdownTimeout,
		router:          engine.Group(""),
	}

	return s
}

func (s *HttpServer) AddRoute(route *rest.Route) error {
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

// Bind reserves the TCP socket the server will later serve on. A port of 0
// lets the OS pick a free one: the effective address is then available through
// the returned listener's Addr method.
func (s *HttpServer) Bind(port uint16) (net.Listener, error) {
	address := fmt.Sprintf(":%d", port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.log.Error("Failed to bind server", slog.String("address", address), slog.Any("error", err))
		return nil, err
	}

	s.log.Info("Server bound", slog.String("address", listener.Addr().String()))

	return listener, nil
}

func (s *HttpServer) Serve(ctx context.Context, listener net.Listener) error {
	address := listener.Addr().String()

	s.log.Info("Starting server", slog.String("address", address))

	srv := &http.Server{
		Handler: s.engine,
	}
	shutdownErrChan := make(chan error, 1)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		shutdownErrChan <- srv.Shutdown(shutdownCtx)
	}()

	err := srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		s.log.Error("Server failed", slog.String("address", address), slog.Any("error", err))
		return err
	}

	if err := <-shutdownErrChan; err != nil {
		s.log.Error("Server shutdown failed", slog.String("address", address), slog.Any("error", err))
		return err
	}

	s.log.Info("Server gracefully shutdown", slog.String("address", address))

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
	e.Use(middleware.RequestId())
	e.Use(middleware.RequestTracer(log))
	e.Use(middleware.RequestLogger())
	e.Use(middleware.ErrorConverter())
	e.Use(middleware.Recover())
}

func buildMiddlewaresForRoute(route *rest.Route) []gin.HandlerFunc {
	out := []gin.HandlerFunc{}

	if route.UseResponseEnvelope() {
		out = append(out, middleware.ResponseEnvelope())
	}

	return out
}
