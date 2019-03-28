package shorty

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	ocpromexp "go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// Shorty main application component.
type Shorty struct {
	config      *Config
	engine      *gin.Engine
	exporter    *ocpromexp.Exporter
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
		Handler: &ochttp.Handler{
			Handler: s.engine,
			GetStartOptions: func(r *http.Request) trace.StartOptions {
				startOptions := trace.StartOptions{}
				if r.URL.Path == "/metrics" {
					startOptions.Sampler = trace.NeverSample()
				}
				return startOptions
			},
		},
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
	s.engine.GET("/", s.handler.Slash)
	s.engine.POST("/shorty/:short", s.handler.Create)
	s.engine.GET("/shorty/:short", s.handler.Read)
	s.engine.GET("/metrics", gin.HandlerFunc(func(c *gin.Context) {
		s.exporter.ServeHTTP(c.Writer, c.Request)
	}))
}

func (s *Shorty) registerExporters() {
	view.RegisterExporter(s.exporter)

	if err := view.Register(
		ochttp.ServerRequestCountView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerLatencyView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerResponseCountByStatusCode,
	); err != nil {
		log.Fatalf("Error on registering metrics: '%s'", err)
	}
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

	s := &Shorty{config: config, engine: gin.Default(), stopChan: make(chan os.Signal, 1)}

	if s.exporter, err = ocpromexp.NewExporter(ocpromexp.Options{
		Registry: prometheus.DefaultGatherer.(*prometheus.Registry),
	}); err != nil {
		return nil, err
	}
	s.registerExporters()

	if persistence, err = NewPersistence(config); err != nil {
		return nil, err
	}
	s.handler = NewHandler(persistence)

	return s, nil
}
