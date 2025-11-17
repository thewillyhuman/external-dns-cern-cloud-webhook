# Makefile for the ExternalDNS CERN Cloud Webhook.

# --- Variables ---
# These variables are used to make the Makefile more maintainable.
# If the binary name, main package path, or image name changes, you only need to
# update them here.

# The name of the binary to be built.
BINARY_NAME=external-dns-cern-cloud-webhook
# The path to the main package.
CMD_PATH=./cmd/webhook
# The name of the Docker image to be built.
IMAGE_NAME=external-dns-cern-cloud-webhook
# The maximum allowed size for the Docker image in bytes (2MB).
MAX_IMAGE_SIZE_BYTES=2097152

# --- Targets ---

# The .PHONY directive tells make that these are not files.
# This is important to avoid conflicts with files of the same name.
.PHONY: all build build-image test-image

# The 'all' target is the default target that is run when you execute 'make'.
all: build

# The 'build' target compiles the Go application into a static binary.
build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Binary '$(BINARY_NAME)' created."

# The 'build-image' target builds the Docker image for the application.
build-image:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME) .
	@echo "Docker image '$(IMAGE_NAME)' built."

# The 'test-image' target tests the Docker image.
# It first builds the image, then checks that it was created successfully and
# that its size does not exceed the defined limit.
test-image: build-image
	@echo "--- Testing Docker Image ---"
	@IMAGE_ID=$$(docker images -q $(IMAGE_NAME)); \
	if [ -z "$$IMAGE_ID" ]; then \
		echo "ERROR: Docker image '$(IMAGE_NAME)' not found."; \
		exit 1; \
	fi
	@IMAGE_SIZE=$$(docker inspect -f '{{.Size}}' $(IMAGE_NAME)); \
	echo "Image size: $$IMAGE_SIZE bytes."; \
	if [ $$IMAGE_SIZE -gt $(MAX_IMAGE_SIZE_BYTES) ]; then \
		echo "ERROR: Image size ($$IMAGE_SIZE bytes) exceeds limit of $(MAX_IMAGE_SIZE_BYTES) bytes."; \
		exit 1; \
	else \
		echo "Image size is within the 2MB limit."; \
	fi
	@echo "Image test passed successfully."

