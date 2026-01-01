# MailNull

A high-performance, concurrent email verification engine built with Go. Provides deep email validation through syntax checking, disposable domain detection, DNS verification, and SMTP handshake validation.

## Architecture

MailNull is a Go/Gin backend service featuring:

- Concurrent worker pool architecture for batch email processing
- Multi-layer verification pipeline
- Three operating modes (LIVE, LITE, MOCK) for different deployment scenarios
- RESTful API with JSON responses

## Email Verification Pipeline

The engine performs verification in multiple stages:

1. **Syntax Validation** - RFC 5322 compliant regex pattern matching
2. **Disposable Domain Check** - Fast lookup against known disposable email providers
3. **DNS MX Record Lookup** - Verifies domain has valid mail servers
4. **SMTP Handshake** - Connects to mail server and validates recipient (Port 25 required)
5. **Catch-All Detection** - Tests random addresses to detect catch-all domains

## Operating Modes

### LIVE Mode (Production)

Full verification including SMTP handshake. Requires Port 25 access.

```bash
go run main.go
```

### LITE Mode (Local Development)

Syntax + DNS validation only. Use when Port 25 is blocked by ISP/firewall.

```bash
MODE=LITE go run main.go
```

### MOCK Mode (Testing)

Returns deterministic mock data. No network calls.

```bash
MODE=MOCK go run main.go
```

## API Endpoints

### POST /v1/mailnull/verify

Verify single or multiple emails concurrently.

**Single Email:**

```bash
curl -X POST http://localhost:8080/v1/mailnull/verify \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

**Batch Verification:**

```bash
curl -X POST http://localhost:8080/v1/mailnull/verify \
  -H "Content-Type: application/json" \
  -d '{"emails": ["user1@example.com", "user2@example.com"]}'
```

**Response (Single Email - Returns Object):**

```json
{
  "email": "user@example.com",
  "is_valid_format": true,
  "is_disposable_email": false,
  "deliverability": "DELIVERABLE",
  "quality_score": 0.7,
  "provider": "example.com",
  "timestamp": "2025-12-29T23:38:00Z"
}
```

**Response (Multiple Emails - Returns Array in Results):**

```json
{
  "results": [
    {
      "email": "user1@example.com",
      "is_valid_format": true,
      "is_disposable_email": false,
      "deliverability": "DELIVERABLE",
      "quality_score": 0.7,
      "provider": "example.com",
      "timestamp": "2025-12-29T23:38:00Z"
    },
    {
      "email": "user2@example.com",
      "is_valid_format": true,
      "is_disposable_email": false,
      "deliverability": "DELIVERABLE",
      "quality_score": 0.7,
      "provider": "example.com",
      "timestamp": "2025-12-29T23:38:01Z"
    }
  ]
}
```

### GET /health

Service health check.

```bash
curl http://localhost:8080/health
```

## Deliverability Status

- **DELIVERABLE** - Email passed all verification stages
- **UNDELIVERABLE** - Invalid syntax, no MX records, or hard bounce
- **RISKY** - Catch-all domain, greylisted, or network timeout
- **UNKNOWN** - Verification incomplete

## Quality Score Algorithm

Weighted scoring based on verification stages:

- Syntax validation: 10%
- MX record presence: 20%
- SMTP RCPT success: 50%
- Non-disposable domain: 20%

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Access to Port 25 (for LIVE mode SMTP verification)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/Veri5ied/mailnull.git
cd mailnull
```

2. Install Go dependencies:

```bash
go mod download
```

3. Run the service:

```bash
go run main.go
```

The API server will start on `http://localhost:8080` by default.

## Development

### Running the Server

```bash
go run main.go
```

### Building for Production

```bash
go build -o mailnull main.go
./mailnull
```

### Testing

```bash
go test ./...
```

## Port 25 Requirement

Full SMTP verification requires outbound Port 25 access. This port is commonly blocked by:

- Residential ISPs (spam prevention)
- Cloud providers (AWS, GCP, Azure by default)
- Corporate firewalls

**For Local Development:** Use LITE mode  
**For Production:** Deploy to VPS with Port 25 access (DigitalOcean, Linode, Vultr, Hetzner)

## Configuration

Environment variables:

- `PORT` - API server port (default: 8080)
- `MODE` - Operating mode: LIVE, LITE, or MOCK (default: LIVE)
- `LOG_LEVEL` - Logging verbosity (default: INFO)

## Technology Stack

- **Go 1.25+** - Primary language
- **Gin** - Web framework for HTTP routing and middleware
- **Standard Library** - net/smtp, log/slog for core functionality
- **Worker Pool** - Concurrent batch processing pattern

## Project Structure

```
mailnull/
├── main.go              # Application entry point
├── handlers/            # HTTP request handlers
│   └── verify.go        # Email verification endpoint handler
├── internal/            # Internal packages
│   ├── config/          # Configuration management
│   │   └── config.go
│   ├── logger/          # Logging configuration
│   │   └── logger.go
│   └── verifier/        # Email verification engine
│       ├── types.go     # Data structures and models
│       ├── verifier.go  # Core verification logic
│       └── worker.go    # Worker pool implementation
├── go.mod               # Go module dependencies
└── go.sum               # Dependency checksums
```

## Performance

- Concurrent worker pool handles batch requests efficiently
- Non-blocking I/O for SMTP connections
- Configurable timeout settings
- Optimized for high-throughput email verification
