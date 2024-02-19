# Stage 1: Build the Go application
FROM golang:1.21.4-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Change to the directory containing main.go
WORKDIR /app/cmd/signalapp

# Build the Go app
# CGO_ENABLED=0: This environment variable disables CGo, which is the Go package that 
# facilitates calling C code from Go. Setting CGO_ENABLED=0 ensures that the Go compiler
# does not link any C library dynamically. This is crucial for creating a statically linked
# binary that does not depend on external C libraries at runtime, making the binary more portable
# across different environments, especially within minimal Docker images that don’t include these
# C libraries.

# GOOS=linux: This tells the Go compiler to build the application for Linux operating systems. 
# Since Docker containers run on a Linux kernel, specifying GOOS=linux ensures that the compiled binary
# is compatible with the Linux environment inside the container, regardless of the host operating system.

# -a: This flag forces the Go compiler to rebuild all dependencies of the application, including standard
# library packages. This ensures that everything is compiled specifically for the target environment specified
# by CGO_ENABLED and GOOS. It's a way to avoid potential issues with cached builds that might not be compatible
# with the target environment.

# -installsuffix cgo: This adds a custom suffix to the path where the compiled packages are stored. Since CGo is
# disabled, adding a suffix like cgo differentiates these builds in the cache. This prevents any conflicts with other
# builds in the cache that might have CGo enabled, ensuring that your dependencies are consistently compiled with CGo disabled.

# -o myapp: This option specifies the name of the output binary. -o stands for "output," and myapp is the chosen name for the
# compiled binary. This means that the go build command will produce an executable named myapp.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# Stage 2: Use a smaller base image
FROM alpine:latest  

# Installs ca-certificates within the Alpine image, allowing your application to make HTTPS requests.
# The --no-cache option prevents the package manager cache from being stored in the final image, 
# reducing size.
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/cmd/signalapp/myapp .

EXPOSE 8080

# Command to run the executable
CMD ["./myapp"]