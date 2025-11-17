# Dockerfile for external-dns-cern-cloud-webhook

# --- Build Stage ---
# This stage is responsible for compiling the Go application into a static binary.
# We use a specific version of the golang:alpine image to ensure a reproducible build environment.
# The 'builder' alias is used to reference this stage in the final stage.
FROM golang:1.24.4-alpine AS builder

# Set the working directory inside the container.
# This is where the subsequent commands will be executed.
WORKDIR /app

# Copy the go.mod and go.sum files first.
# This leverages Docker's layer caching. If these files haven't changed, Docker will
# reuse the cached layer from the next step, speeding up subsequent builds.
COPY go.mod go.sum ./

# Download the Go module dependencies.
# This is done before copying the rest of the source code to further optimize the build cache.
RUN go mod download

# Copy the rest of the application's source code into the container.
COPY . .

# Install UPX, a utility for compressing executables.
# This is used to significantly reduce the size of the final binary.
# The --no-cache flag is used to avoid storing the package index, keeping the layer small.
RUN apk add --no-cache upx

# Compile the Go application.
# - CGO_ENABLED=0: Disables CGO, which is necessary for creating a static binary.
# - GOOS=linux: Specifies that the binary should be compiled for the Linux operating system.
# - ldflags="-s -w": Strips debugging information from the binary, reducing its size.
# - trimpath: Removes all file system paths from the resulting executable, improving build reproducibility.
# -o app: Specifies the output file name for the compiled binary.
# ./cmd/webhook: Specifies the main package to compile.
# The '&& upx --best --lzma app' command then compresses the compiled binary using UPX
# with the best compression settings.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o app ./cmd/webhook && upx --best --lzma app

# --- Final Stage ---
# This stage is responsible for creating the final, minimal container image.
# We use the 'scratch' image, which is an empty image, as the base.
# This results in the smallest possible image size and a reduced attack surface,
# as it contains only our application and nothing else.
FROM scratch

# Set the working directory inside the container.
WORKDIR /root/

# Copy the compiled and compressed binary from the 'builder' stage into the final image.
# This is the only file that will be included in the final image.
COPY --from=builder /app/app .

# Set the default command to run when the container starts.
# This will execute our application's binary.
CMD ["./app"]
