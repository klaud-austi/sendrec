# SendRec

Privacy-first screen recording with EU-only storage.

## Quick Start

```bash
# Clone and start
git clone https://github.com/sendrec/sendrec.git
cd sendrec
docker-compose up -d

# Access at http://localhost:8080
```

## Features

- **Landing Page** - Marketing site at `/`
- **Waitlist API** - Join at `POST /waitlist`
- **Admin Panel** - View subscribers at `/admin`
- **Health Check** - `GET /health`

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Landing page |
| POST | `/waitlist` | Join waitlist (JSON: `{"email":"user@example.com"}`) |
| GET | `/admin` | View waitlist subscribers |
| GET | `/health` | Health check |

## Development

```bash
# Run locally
go run .

# Build binary
go build -o sendrec .

# Run binary
./sendrec
```

## Deployment

See [DEPLOY.md](../DEPLOY.md) for complete deployment instructions.

## License

AGPL-3.0
