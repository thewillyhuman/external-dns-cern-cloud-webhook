# ExternalDNS CERN Cloud Webhook

This project provides a webhook provider for [ExternalDNS](https://github.com/kubernetes-sigs/external-dns)
to manage DNS records within the CERN cloud infrastructure. It acts as a bridge
between Kubernetes and the CERN DNS system, allowing you to control DNS records
declaratively through Kubernetes resources.

## Repository Structure

The repository is organized into several packages to maintain a clean and
modular architecture:

-   `/cmd/webhook`: The main entry point for the application.
-   `/pkg/config`: Defines the application's configuration structure.
-   `/pkg/webhook`: Implements the HTTP server that listens for requests from
    ExternalDNS.
-   `/provider`: Contains the core logic for interacting with the CERN Cloud DNS
    service.
-   `/internal/log`: Provides a structured logging interface for the application.
-   `/Dockerfile`: Defines the multi-stage build process for creating a minimal
    and secure container image.
-   `/Makefile`: Contains helper targets for common development tasks like
    building and testing.

## Getting Started

### Building the Project

A `Makefile` is provided to simplify the build process.

-   **To build the binary:**
    ```bash
    make build
    ```

-   **To build the Docker image:**
    ```bash
    make build-image
    ```

-   **To test the Docker image (including a size check):**
    ```bash
    make test-image
    ```

## License

This project is licensed under the BSD 3-Clause License. See the [LICENSE](LICENSE)
file for details.

The core requirement of this license is that if you use this software, you must
include the original copyright notice in your distribution. This ensures that
the original authors are always credited for their work.
