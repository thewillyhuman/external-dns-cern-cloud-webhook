// Package webhook provides the implementation of the ExternalDNS webhook server.
//
// This package is responsible for setting up and running the HTTP server that listens for
// requests from ExternalDNS. It defines the HTTP routes and handlers, and delegates the
// business logic to the provider package.
// The separation of the server logic from the main application entry point and the provider
// logic promotes a clean and modular architecture.
package webhook

import (
	"fmt"
	"net/http"
	"os"

	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/log"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/pkg/config"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/provider"
)

// Server is the main struct for the webhook server.
//
// It holds a reference to the provider, which contains the business logic for
// interacting with the DNS service, and the application configuration.
// This design was chosen to keep the server's responsibilities focused on handling
// HTTP requests and routing, while the provider handles the specifics of the DNS
// provider.
type Server struct {
	provider *provider.Provider
	config   *config.Config
}

// NewServer creates a new instance of the webhook server.
//
// It takes a provider and a configuration object as input and returns a new Server.
// This is the preferred way to create a new server, as it ensures that the server
// is properly initialized with all its dependencies.
func NewServer(p *provider.Provider, cfg *config.Config) *Server {
	return &Server{provider: p, config: cfg}
}

// Run starts the webhook server and begins listening for incoming requests.
//
// This method sets up the HTTP routes and handlers, and then starts the HTTP server.
// It is a blocking call that will run until the application is terminated.
// The routing is designed to match the ExternalDNS webhook provider specification,
// ensuring compatibility with ExternalDNS.
func (s *Server) Run() {
	// The /records endpoint is special as it handles both GET and POST requests.
	// A dedicated handler is used to switch on the request method and delegate to the
	// appropriate provider method. This is a clean way to handle multiple methods
	// on the same endpoint.
	recordsHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.provider.Records(w, r)
		case http.MethodPost:
			s.provider.ApplyChanges(w, r)
		default:
			// If the request method is not GET or POST, return a 405 Method Not Allowed error.
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}

	// Register the HTTP handlers for the various endpoints.
	// Each handler is a method on the provider, which keeps the business logic
	// separate from the server logic.
	http.HandleFunc("/", s.provider.Negotiate)
	http.HandleFunc("/records", recordsHandler)
	http.HandleFunc("/adjustendpoints", s.provider.AdjustEndpoints)
	http.HandleFunc("/healthz", s.provider.Healthz)

	// Create the server address from the configured listen address and port.
	addr := fmt.Sprintf("%s:%d", s.config.ListenAddress, s.config.ListenPort)

	// Create a new http.Server instance.
	// This allows for more control over the server's configuration in the future,
	// such as setting timeouts or enabling TLS.
	server := &http.Server{
		Addr: addr,
	}

	// Start the HTTP server and log a message to indicate that it is running.
	log.GlobalLogger.Info("Listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		// If the server fails to start, log the error and exit the application.
		log.GlobalLogger.Error("failed to start server: %v", err)
		os.Exit(1)
	}
}
