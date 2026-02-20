package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/tainj/distributed_calculator2/internal/auth"
	"github.com/tainj/distributed_calculator2/internal/transport/grpc/handlers"
	"github.com/tainj/distributed_calculator2/internal/transport/grpc/middlewares"
	client "github.com/tainj/distributed_calculator2/pkg/api"
	"github.com/tainj/distributed_calculator2/pkg/logger"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Server with grpc and rest
type Server struct {
	grpcServer *grpc.Server
	restServer *http.Server
	listener   net.Listener
}

// New creates a new grpc + rest server
func New(ctx context.Context,
	port, restPort int,
	service handlers.Service,
	jwtService auth.JWTService) (*Server, error) {

	// get logger from context
	loggerFromCtx := logger.GetLoggerFromCtx(ctx)

	lis, err := net.Listen("tcp", fmt.Sprintf("api:%d", port))
	if err != nil {
		loggerFromCtx.Error(ctx, "failed to listen", "error", err)
	}

	// configure grpc with logging
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				// request start
				loggerFromCtx.Info(ctx, "grpc request started",
					"method", info.FullMethod,
					"request_type", fmt.Sprintf("%T", req))

				// call handler
				resp, err := handler(ctx, req)

				// result
				if err != nil {
					loggerFromCtx.Error(ctx, "grpc request failed",
						"method", info.FullMethod,
						"error", err.Error())
				} else {
					loggerFromCtx.Info(ctx, "grpc request completed",
						"method", info.FullMethod)
				}

				return resp, err
			},
		),
	}

	grpcServer := grpc.NewServer(opts...)
	calculatorService := handlers.NewCalculatorService(service)
	client.RegisterCalculatorServer(grpcServer, calculatorService)

	// rest gateway
	restMux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.Pairs()
			// pass through required headers
			for key, values := range r.Header {
				if key == "Authorization" || key == "Content-Type" {
					md.Set(key, values...)
				}
			}
			return md
		}),
	)

	if err := client.RegisterCalculatorHandlerServer(context.Background(), restMux, calculatorService); err != nil {
		return nil, err
	}

	// middleware for rest with logging
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// http request start
			loggerFromCtx.Info(ctx, "http request started",
				"method", r.Method,
				"url", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent())

			// to catch status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// call handler
			next.ServeHTTP(wrapped, r)

			// completion
			loggerFromCtx.Info(ctx, "http request completed",
				"method", r.Method,
				"url", r.URL.Path,
				"status", wrapped.statusCode)
		})
	}

	// all middleware together
	finalHandler := middlewares.Apply(
		loggingMiddleware(restMux),
		middlewares.LoggerProvider("calculator-gateway"),
		middlewares.AuthMiddleware(jwtService),
		middlewares.Logging(),
	)

	// cors - allow frontend
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173", // vite
			"http://localhost:3000", // create-react-app
		},
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS", "PUT", "DELETE",
		},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "Content-Length", "Authorization", "X-Requested-With",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}).Handler(finalHandler)

	// create http server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", restPort),
		Handler: corsHandler,
	}

	return &Server{grpcServer, httpServer, lis}, nil
}

// helper structure - to get response status
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Start starts grpc and rest servers
func (s *Server) Start(ctx context.Context) error {
	l := logger.GetLoggerFromCtx(ctx)
	eg := errgroup.Group{}

	eg.Go(func() error {
		l.Info(ctx, "starting grpc server", "port", s.listener.Addr().(*net.TCPAddr).Port)
		return s.grpcServer.Serve(s.listener)
	})

	eg.Go(func() error {
		l.Info(ctx, "starting rest server", "port", s.restServer.Addr)
		return s.restServer.ListenAndServe()
	})

	return eg.Wait()
}

// Stop stops servers
func (s *Server) Stop(ctx context.Context) error {
	s.grpcServer.GracefulStop()
	l := logger.GetLoggerFromCtx(ctx)
	if l != nil {
		l.Info(ctx, "grpc server stopped")
	}
	return s.restServer.Shutdown(ctx)
}
