# Project Overview

This project is a webhook provider for ExternalDNS that manages DNS records at
CERN. It is a Go application that implements the ExternalDNS webhook provider
interface. The application is designed with a clean and modular architecture,
separating the command-line interface, the web server, and the provider logic
into distinct packages.

# Technologies

*   **Go:** The application is written in Go.
*   **zerolog:** Used for structured logging.
*   **pflag & viper:** Used for configuration management from command-line flags
    and environment variables.

# Architecture

The application is divided into the following packages:

*   `cmd/webhook`: The main entry point of the application. It is responsible
    for parsing command-line flags and environment variables, initializing the
    configuration and logger, and starting the webhook server.
*   `pkg/config`: Defines the application's configuration structure.
*   `pkg/webhook`: Implements the HTTP server that listens for requests from
    ExternalDNS. It routes requests to the appropriate provider methods.
*   `provider`: Implements the ExternalDNS webhook provider interface. This is
    where the business logic for managing DNS records at CERN will be
    implemented.
*   `internal/log`: Defines a logging interface and provides a zerolog
    implementation.

# Developing Code

Before writing new code:
1.  Run `/memory refresh` to update any change done outside the gemini-cli.

After writing any code:
1.  Format all code to code conventions or what it is stated in the
    GEMINI.md file.
2.  Run test suite.
3.  Test docker image via `make test-image`.
4.  If all success, be sure that you have documented all code properly.

# Testing

**TODO:** Add instructions for running the tests once they are implemented.

# Building and Running

## Building

### Building the binary

To build the application, run the following command from the root of the project:

```bash
go build -o external-dns-cern-cloud-webhook ./cmd/webhook
```

### Building the Docker image

To build the Docker image, run the following command from the root of the
project:

```bash
docker build -t external-dns-cern-cloud-webhook .
```

## Running

To run the application, you can use the following command:

```bash
./external-dns-cern-cloud-webhook --log-level debug
```

The application can be configured using command-line flags or environment
variables. For a full list of options, run:

```bash
./external-dns-cern-cloud-webhook --help
```

## Makefile Targets

A `Makefile` is provided to simplify common development tasks.

*   `make build`: Compiles the Go application into a static binary named
    `external-dns-cern-cloud-webhook`.
*   `make build-image`: Builds the Docker image for the application with the
    tag `external-dns-cern-cloud-webhook`.
*   `make test-image`: Builds the Docker image and then runs a test to ensure
    it was created successfully and that its size does not exceed 2MB.

# Development Conventions

## Coding Style

The project follows the standard Go coding style. Use `go fmt` to format your
code before committing. All non-Go files (e.g., Markdown, YAML) should have a
maximum line length of 80 characters.

## Logging

The project uses the `zerolog` library for structured logging. A global logger
is available in the `internal/log` package.

## Configuration

Configuration is managed using the `pflag` and `viper` libraries. Default
values are defined in `cmd/webhook/options.go`. Configuration can be provided
via command-line flags or environment variables.

## Testing

**TODO**