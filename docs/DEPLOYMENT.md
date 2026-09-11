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

## Reverse Proxy

The `backend` and `frontend` services join an external Docker network named `caddy`
in addition to the project's default network, and use fixed container names
(`minicord-backend`, `minicord-frontend`) so a separately managed reverse proxy can
reach them by name. Create the network once if it does not already exist:

```bash
docker network create caddy
```

Example reverse proxy site block (Caddy) that terminates TLS and forwards
application API traffic to the backend and everything else to the frontend:

```text
your-domain.example {
    reverse_proxy /api/* minicord-backend:8080
    reverse_proxy minicord-frontend:80
}
```

The reverse proxy is deployed and managed outside this repository; only the shared
Docker network and container names are defined here.
