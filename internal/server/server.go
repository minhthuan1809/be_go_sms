package server

import (
	"log"
	"net/http"

	"myproject/internal/handler"
)

// Server represents the HTTP server
type Server struct {
	addr    string
	handler *handler.SMSHandler
}

// NewServer creates a new server instance
func NewServer(addr string, version string) *Server {
	return &Server{
		addr:    addr,
		handler: handler.NewSMSHandler(version),
	}
}

// SetupRoutes configures all HTTP routes
func (s *Server) SetupRoutes() {
	http.HandleFunc("/", s.handler.HandleRoot)
	http.HandleFunc("/send", s.handler.HandleSendSMS)
	http.HandleFunc("/health", s.handler.HandleHealth)
	http.HandleFunc("/port/status", s.handler.HandlePortStatus)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.SetupRoutes()

	log.Printf("SMS Gateway API v%s listening on %s", s.handler.Version(), s.addr)
	log.Printf("Available endpoints:")
	log.Printf("  POST /send - Send SMS")
	log.Printf("  GET /health - Health check")
	log.Printf("  GET /port/status?port=<port> - Check port status")

	return http.ListenAndServe(s.addr, nil)
}
