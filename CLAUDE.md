# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gomments is a simple commenting system for the personal website less.coffee. It's built with Go using Gin for the web framework and SQLite for data storage. The system allows anonymous commenting with optional signature-based identification using tripcodes.

## Build and Development Commands

### Building and Running
- `make build` - Build Docker image
- `make run` - Build and run the application with Docker (requires PORT and BASE_URL env vars)
- `make run-release` - Run in production mode with GIN_MODE=release
- `go build ./cmd/app` - Build the Go binary directly
- `go run ./cmd/app/main.go` - Run directly with Go (requires PORT env var)

### Testing
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test ./... -run TestName` - Run specific test

### Other Go Commands
- `go mod tidy` - Clean up dependencies
- `go fmt ./...` - Format code
- `go vet ./...` - Run Go vet for static analysis

## Architecture

### Core Structure
- `cmd/app/main.go` - Application entry point with environment setup and server initialization
- `service.go` - Core business logic service layer with comment operations
- `dao.go` - Database access layer with SQLite queries
- `internal/router.go` - HTTP route definitions and handlers
- `internal/database.go` - Database initialization and migration setup
- `error.go` - Custom error handling types

### Key Components
- **Service Layer**: `gomments.Service` handles business logic for getting replies, submitting comments, and generating statistics
- **Database Layer**: Uses `sqlx` for SQLite operations with automatic migrations
- **API Layer**: Gin-based REST API with CORS support for `less.coffee` domain
- **Authentication**: Tripcode-based signatures for optional user identification

### Database
- SQLite database with migrations in `migrations/` directory
- Automatic migration on startup
- Primary table: `reply` with fields for comments, articles, signatures, and metadata

### Environment Variables
- `PORT` (required) - Server port
- `BASE_URL` (optional) - Base URL for API routes
- `GIN_MODE` (optional) - Set to "release" for production

### API Endpoints
- `GET /ping` - Health check
- `GET /articles/:article/replies` - Get comments for an article
- `POST /articles/:article/replies` - Submit new comment
- `GET /articles/replies/stats` - Get comment statistics for multiple articles

### Testing
- Uses testify/require for assertions
- Test fixtures create temporary SQLite databases
- Tests run in parallel with cleanup
- Located in `service_test.go` and `fixture_test.go`