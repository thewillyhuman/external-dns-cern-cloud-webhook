// Package main is the entry point for the ExternalDNS CERN Cloud Webhook.
//
// This application starts a webhook server that listens for requests from ExternalDNS.
// It is responsible for initializing the configuration, setting up logging,
// creating a provider instance, and starting the webhook server.
// The layered architecture separates the command-line logic from the server and provider implementations,
// promoting modularity and maintainability.
package main

import (
	"os"

	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/log"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/pkg/webhook"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/provider"
)

// main is the entry point of the application.
//
// It performs the following steps:
//  1. Loads the application configuration from command-line flags and environment variables.
//     This is a critical first step as it dictates the behavior of the entire application.
//     If configuration loading fails, the application will exit with a non-zero status code.
//  2. Sets up the global logger with the log level specified in the configuration.
//     This ensures that all subsequent log messages are filtered and formatted correctly.
//     A default logger is used for logging configuration errors that occur before the
//     final logger is configured.
//  3. Creates a new provider instance, passing in the loaded configuration.
//     The provider is responsible for the business logic of interacting with the
//     CERN Cloud DNS service.
//  4. Creates a new webhook server, passing in the provider and configuration.
//     The server is responsible for handling HTTP requests from ExternalDNS and
//     delegating them to the provider.
//  5. Starts the webhook server, which begins listening for incoming requests.
//     This is a blocking call, and the application will continue to run until the
//     server is stopped.
func main() {
	// Load configuration from flags and environment variables.
	// We use a dedicated configuration package to keep this logic separate from the main application logic.
	// This makes it easier to manage configuration and add new options in the future.
	cfg, err := loadConfig()
	if err != nil {
		// If configuration loading fails, we need to log the error and exit.
		// Since the global logger is not yet configured, we create a temporary one with the default log level.
		log.GlobalLogger = log.NewLogger(log.DefaultLogLevel)
		log.GlobalLogger.Error("failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Set up the global logger based on the configured log level.
	// The log level is parsed from a string to a log.Level type.
	// If the log level is invalid, a warning is logged, and the default log level is used.
	logLevel, ok := log.LevelFromString(cfg.LogLevel)
	if !ok {
		log.GlobalLogger = log.NewLogger(log.DefaultLogLevel)
		log.GlobalLogger.Warn("invalid log level '%s', using default '%s'", cfg.LogLevel, log.LevelNames[log.DefaultLogLevel])
		logLevel = log.DefaultLogLevel
	}
	log.GlobalLogger = log.NewLogger(logLevel)

	// Create a new provider instance.
	// The provider encapsulates the logic for interacting with the CERN Cloud DNS service.
	// It is initialized with the application configuration, which it uses to configure its own behavior.
	p := provider.NewProvider(cfg)

	// Create a new webhook server.
	// The server is responsible for handling HTTP requests from ExternalDNS.
	// It is initialized with the provider and the application configuration.
	srv := webhook.NewServer(p, cfg)

	// Start the webhook server.
	// This is a blocking call that will run until the application is terminated.
	srv.Run()
}
