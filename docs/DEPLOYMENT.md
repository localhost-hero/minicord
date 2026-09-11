# Local Deployment

## Prerequisites

- Docker Desktop with Docker Compose;
- Go 1.24 or newer for local backend development;
- Node.js 22 or newer and npm for local frontend development.

## Configuration

Copy the example configuration before starting services:

```bash
cp .env.example .env
```

Replace placeholder LiveKit credentials before using a non-development deployment. Never commit `.env`.

`LIVEKIT_URL` is returned to the browser as part of the room token response, so it
must be an address the client can reach directly (for example `ws://localhost:7880`
for local development, or a public `wss://` LiveKit hostname in production) rather
than an internal container hostname.

## Start the Stack

From the repository root:

```bash
docker compose up --build
```

Services:

- frontend: `http://localhost:5173`;
- backend health: `http://localhost:8080/health`;
- LiveKit signaling: `ws://localhost:7880`;
- LiveKit RTC TCP: port `7881`;
- LiveKit RTC UDP: port `7882`.

The current Compose file uses LiveKit development mode for local work. Do not use development credentials or unsecured local endpoints in production.

## Stop the Stack

```bash
docker compose down
```

## Local Checks

Backend:

```bash
cd backend
go test ./...
go vet ./...
```

Frontend:

```bash
cd frontend
npm install
npm run build
```

## Production Notes

A production deployment still requires:

- TLS through a reverse proxy;
- real LiveKit API credentials stored as secrets;
- DNS and firewall configuration;
- appropriate LiveKit and TURN ports;
- resource limits and monitoring;
- backups for any persistent application data.
