package shorty

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// Shorty main application component.
type Shorty struct {
	config      *Config
	router      *mux.Router
	handler     *Handler
	persistence *Persistence
	stopChan    chan os.Signal
}

// httpServer uses configuration to spin up a new http server, and start serving content until os
// signal is sent.
func (s Shorty) httpServer() {
	server := &http.Server{
		Addr:         s.config.Address,
		ReadTimeout:  time.Duration(s.config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.IdleTimeout) * time.Second,
		Handler:      s.router,
	}

	go func() {
		log.Printf("Listening on '%s'", s.config.Address)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error: '%s'", err)
			s.stopChan <- os.Kill
		}
	}()

	// block until os signal is sent
	<-s.stopChan
}

// setUpRoutes define how the routes are configured for this application.
func (s *Shorty) setUpRoutes() {
	s.router.Methods("GET").Path("/metrics").Name("metrics").
		Handler(prometheus.Handler())
	s.router.Methods("GET").Path("/").Name("slash").
		Handler(Logger(prometheus.InstrumentHandlerFunc("slash", s.handler.Slash()), "slash"))
	s.router.Methods("GET").Path("/{short}").Name("read").
		Handler(Logger(prometheus.InstrumentHandlerFunc("read", s.handler.Read()), "read"))
	s.router.Methods("POST").Path("/{short}").Name("create").
		Handler(Logger(prometheus.InstrumentHandlerFunc("create", s.handler.Create()), "create"))
}

// Run creates the runtime instance, add routes and start http-server.
func (s Shorty) Run() error {
	s.setUpRoutes()
	s.httpServer()

	return nil
}

// Shutdown sends os.Kill signal in stop channel.
func (s *Shorty) Shutdown() {
	s.stopChan <- os.Kill
}

// NewShorty new application instance with basic components.
func NewShorty(config *Config) (*Shorty, error) {
	var persistence *Persistence
	var err error

	if persistence, err = NewPersistence(config); err != nil {
		return nil, err
	}

	s := &Shorty{
		config:   config,
		router:   mux.NewRouter().StrictSlash(true),
		handler:  NewHandler(persistence),
		stopChan: make(chan os.Signal, 1),
	}
	return s, nil
}
