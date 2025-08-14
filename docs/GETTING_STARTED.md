# Getting Started with AIOS

Welcome to AIOS (AI-Powered Operating System)! This guide will help you get started with development and deployment.

## Quick Start

### Prerequisites

- **Operating System**: Arch Linux (recommended for development)
- **Go**: Version 1.22 or higher
- **Node.js**: Version 18 or higher
- **Docker**: Latest version with Docker Compose
- **Git**: For version control

### Development Setup

1. **Clone the repository:**
   ```bash
   git clone <your-repo-url> aios
   cd aios
   ```

2. **Run the automated setup:**
   ```bash
   ./scripts/setup-dev.sh
   ```

3. **Start the development environment:**
   ```bash
   make dev
   ```

### Manual Setup (Alternative)

If you prefer manual setup or the automated script doesn't work:

1. **Install dependencies:**
   ```bash
   # Update system
   sudo pacman -Syu

   # Install base tools
   sudo pacman -S base-devel go nodejs npm docker docker-compose postgresql redis

   # Install Go tools
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install github.com/air-verse/air@latest
   ```

2. **Configure services:**
   ```bash
   # Start and enable services
   sudo systemctl enable --now docker postgresql redis

   # Add user to docker group
   sudo usermod -aG docker $USER
   ```

3. **Install project dependencies:**
   ```bash
   go mod download
   cd web && npm install && cd ..
   ```

## Project Structure

```
aios/
├── cmd/                    # Application entrypoints
│   ├── aios-daemon/       # Main system daemon
│   ├── aios-assistant/    # AI assistant service
│   └── aios-desktop/      # Desktop environment
├── internal/              # Core application logic
│   ├── ai/               # AI service implementations
│   ├── system/           # System integration
│   ├── security/         # Security and privacy
│   └── desktop/          # Desktop environment logic
├── pkg/                   # Shared utilities and packages
├── web/                   # Frontend applications
├── configs/               # Configuration files
├── scripts/               # Build and deployment scripts
├── docs/                  # Documentation
└── tests/                 # Test suites
```

## Development Workflow

### Building the Project

```bash
# Build all components
make build

# Build individual components
make build-daemon
make build-assistant
make build-desktop

# Build for Linux
make build-linux
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make benchmark
```

### Code Quality

```bash
# Format code
make format

# Run linter
make lint

# Run go vet
make vet
```

### Development Environment

```bash
# Start development environment
make dev

# Stop development environment
make dev-stop

# View logs
make dev-logs
```

## Configuration

### Environment Variables

Key environment variables for development:

```bash
export AIOS_ENV=development
export AIOS_LOG_LEVEL=debug
export AIOS_CONFIG_PATH=./configs/dev.yaml
export AIOS_DB_HOST=localhost
export AIOS_DB_PORT=5432
export AIOS_REDIS_HOST=localhost
export AIOS_REDIS_PORT=6379
```

### Configuration Files

- `configs/dev.yaml` - Development configuration
- `configs/prod.yaml` - Production configuration
- `configs/test.yaml` - Test configuration

## Services Overview

### AIOS Daemon (Port 8080)

The main system daemon that orchestrates all AI services and system management.

**Key endpoints:**
- `GET /health` - Health check
- `GET /api/v1/system/status` - System status
- `GET /api/v1/resources` - Resource information
- `POST /api/v1/system/optimize` - Trigger optimization

### AI Assistant (Port 8081)

Natural language interface for system interaction.

**Key endpoints:**
- `POST /api/v1/chat` - Chat with AI
- `POST /api/v1/voice` - Voice commands
- `GET /api/v1/ws` - WebSocket connection

### Desktop Environment (Port 8082)

AI-aware desktop environment with intelligent window management.

**Key endpoints:**
- `GET /api/v1/windows` - List windows
- `GET /api/v1/workspaces` - List workspaces
- `GET /api/v1/applications` - List applications

## Monitoring and Observability

### Metrics (Port 9090)

Prometheus metrics are available at:
- `http://localhost:9090/metrics` (daemon)
- `http://localhost:9091/metrics` (assistant)
- `http://localhost:9092/metrics` (desktop)

### Tracing

Jaeger tracing is available at:
- `http://localhost:16686` - Jaeger UI

### Logs

Structured JSON logs are output to stdout and can be viewed with:
```bash
make dev-logs
```

## Database

### PostgreSQL

Development database connection:
- Host: localhost
- Port: 5432
- Database: aios_dev
- User: aios
- Password: aios_password

### Redis

Cache and session storage:
- Host: localhost
- Port: 6379
- Database: 0

## Frontend Development

### Web Interface

The web interface is built with Next.js and TypeScript:

```bash
cd web
npm run dev    # Start development server
npm run build  # Build for production
npm run test   # Run tests
```

### Styling

- **Framework**: Tailwind CSS
- **Components**: Headless UI
- **Icons**: Heroicons, Lucide React
- **Animations**: Framer Motion

## Testing

### Unit Tests

```bash
# Go tests
go test ./...

# Frontend tests
cd web && npm test
```

### Integration Tests

```bash
make test-integration
```

### End-to-End Tests

```bash
make test-e2e
```

## Debugging

### Go Debugging

Use Delve debugger:
```bash
dlv debug ./cmd/aios-daemon
```

### Frontend Debugging

Use browser developer tools or VS Code debugger.

### Docker Debugging

Access running containers:
```bash
docker exec -it aios-daemon-dev /bin/sh
```

## Common Issues

### Port Conflicts

If ports are already in use, modify the configuration in `configs/dev.yaml`.

### Permission Issues

Ensure your user is in the docker group:
```bash
sudo usermod -aG docker $USER
```

### Database Connection Issues

Ensure PostgreSQL is running:
```bash
sudo systemctl status postgresql
```

### Go Module Issues

Clean and rebuild:
```bash
go clean -modcache
go mod download
```

## Next Steps

1. **Explore the codebase** - Start with `cmd/aios-daemon/main.go`
2. **Run the tests** - Ensure everything works
3. **Make changes** - Follow the contributing guidelines
4. **Submit PRs** - Help improve AIOS

## Getting Help

- **Documentation**: Check the `docs/` directory
- **Issues**: Create GitHub issues for bugs
- **Discussions**: Use GitHub discussions for questions
- **Contributing**: See `CONTRIBUTING.md`

## Resources

- [Architecture Documentation](../ARCHITECTURE.md)
- [API Documentation](./api/)
- [Contributing Guidelines](../CONTRIBUTING.md)
- [Roadmap](../ROADMAP.md)

Happy coding with AIOS! 🚀
