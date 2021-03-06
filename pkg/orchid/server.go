package orchid

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/isutton/orchid/pkg/orchid/apiserver"
	"github.com/isutton/orchid/pkg/orchid/config"
	"github.com/isutton/orchid/pkg/orchid/repository"
)

// Options are the server parameters.
type Options struct {
	Address string
}

// Server is the API server.
type Server struct {
	Logger logr.Logger
	Server *http.Server
}

// NewServer creates a new Server using options.
func NewServer(logger logr.Logger, options Options) *Server {
	// TODO: move artificial configuration away;
	config := &config.Config{Username: "postgres", Password: "1", Options: "sslmode=disable"}

	repo := repository.NewRepository(logger, config)
	if err := repo.Bootstrap(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	h := apiserver.NewAPIResourceHandler(logger, repo)
	h.Register(router)

	return &Server{
		Logger: logger.WithName("server"),
		Server: &http.Server{Addr: options.Address, Handler: router},
	}
}

// Start initializes the server without blocking.
func (s *Server) Start(ctx context.Context) error {
	// errChan is used to receive error messages when initializing the server
	errChan := make(chan error)
	defer close(errChan)

	go func() {
		// this goroutine will be executed until ListenAndServe returns
		errChan <- s.Server.ListenAndServe()
	}()

	// wait until either an error or a timeout happen, indicating initialization has been successful
	select {
	case err := <-errChan:
		return err
	case <-time.After(3 * time.Second):
		return nil
	}
}

// Shutdown stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

// AddAPIResourceHandler registers the API server routes in router.
// func AddAPIResourceHandler(logger logr.Logger, crdService apiserver.Model, router *mux.Router) {
// 	// NewAPIResourceHandler(logger, crdService).Register(router)
// }
