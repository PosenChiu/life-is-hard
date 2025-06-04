# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24.3-alpine AS build
WORKDIR /src
# Install build dependencies
RUN apk add --no-cache git
# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy the source code
COPY . .
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/service cmd/service/service.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/service ./service
# Make sure the binary is executable
RUN chmod +x ./service
EXPOSE 8080
CMD ["./service"]
