# Go SDK

This Go SDK provides tools to easily integrate common features into your Go applications, including:

- Initialization of a logger with `slog`
- HTTP middleware to log request and response bodies in JSON
- gRPC interceptors for logging
- JWT token manipulation utilities

## Prerequisites

Make sure you have Go 1.18+ installed on your machine. You can download Go from [https://golang.org/dl/](https://golang.org/dl/).

### Dependencies

This project depends on the following packages:

- `golang.org/x/crypto` : For JWT-related operations
- `github.com/gorilla/mux` : For HTTP routing (if used)
- `github.com/grpc/grpc-go` : For gRPC interceptors
- `golang.org/x/slog` : For structured logging

## Installation

Clone the project and install the dependencies using `go mod`:

```bash
git clone https://github.com/mama165/sdk-go
cd sdk-go
go mod tidy
