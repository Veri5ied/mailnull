# MailNull

A high-performance, concurrent email verification engine built with Go and Next.js. Provides deep email validation through syntax checking, disposable domain detection, DNS verification, and SMTP handshake validation.

## Architecture

This is a Turborepo monorepo containing:

- `apps/api` - Go/Gin backend email verification service
- `apps/web` - Next.js frontend interface
- Concurrent worker pool architecture for batch processing
- Multi-layer verification pipeline

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
cd apps/api
go run main.go
```

### LITE Mode (Local Development)

Syntax + DNS validation only. Use when Port 25 is blocked by ISP/firewall.

```bash
cd apps/api
MODE=LITE go run main.go
```

### MOCK Mode (Testing)

Returns deterministic mock data. No network calls.

```bash
cd apps/api
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

## Development

**Install Dependencies:**

```bash
pnpm install
```

**Run All Services:**

```bash
npm run dev
```

**Run API Only:**

```bash
cd apps/api
go run main.go
```

**Run Frontend Only:**

```bash
cd apps/web
npm run dev
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

**Backend:**

- Go 1.25+
- Gin web framework
- Standard library (net, slog)
- Worker pool concurrency pattern

**Frontend:**

- Next.js 16
- TypeScript
- TailwindCSS

**Infrastructure:**

- Turborepo monorepo
- CORS enabled for localhost:3000
