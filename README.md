# Self-Hosted Audio Web

A lightweight self-hosted web application for group voice calls.

The project is designed for a `user -> server -> user` architecture. Audio must be routed through the owner's infrastructure; clients must not establish direct peer-to-peer audio connections.

> Status: early development. The initial Go and React/Vite scaffold is available. LiveKit runtime validation requires Docker Desktop.

## Planned Architecture

```text
Browser A ──┐
            ├──> LiveKit SFU ──> Browser participants
Browser B ──┘
       ^
       |
Go application backend
```

The Go backend will manage application concerns:

- authentication and authorization;
- users and rooms;
- room access checks;
- temporary LiveKit token issuance;
- application health and configuration APIs.

LiveKit will manage media-session signaling and audio routing. coturn may be used as an additional relay for restrictive networks. Direct peer-to-peer fallback is not part of the architecture.

## Planned Stack

- Backend: Go
- Frontend: React, TypeScript, and Vite
- Media: WebRTC through LiveKit SFU
- Audio codec: Opus
- Relay: coturn when required by the deployment
- Local deployment: Docker Compose
- Transport security: TLS through a reverse proxy

## MVP Scope

The first working vertical slice will provide:

- a Go backend with a health endpoint;
- a React frontend shell;
- local LiveKit development service;
- server-side room access validation;
- temporary LiveKit tokens;
- joining and leaving a voice room;
- mute and unmute;
- participant list and active-speaker state;
- reconnecting and failure states;
- tests for configuration, authorization, and call state transitions.

Chat, recording, user profiles, production deployment, and advanced moderation are out of scope for the first MVP.

## Repository Structure

```text
.github/
  copilot-instructions.md
  pull_request_template.md
backend/                 # Go application
frontend/                # React application
docs/                    # architecture and operational documentation
CONTRIBUTING.md          # contribution and review workflow
LICENSE
README.md
```

## Local Development

Install Docker Desktop, then copy the example configuration and start the local stack:

```bash
cp .env.example .env
docker compose up --build
```

Open the frontend at `http://localhost:5173` and check the backend at `http://localhost:8080/health`.

For focused local checks:

```bash
cd backend && go test ./... && go vet ./...
cd ../frontend && npm install && npm run build
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for service ports and deployment notes.

## Development Status

The initial development task is tracked in [Issue #1](https://github.com/localhost-hero/minicord/issues/1).

The implementation is added incrementally through focused branches and Pull Requests. The first scaffold is complete; authentication, token issuance, and the real call flow are the next implementation slices.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before making changes. The project uses the following workflow:

```text
Issue -> branch -> implementation -> tests -> Pull Request -> review -> merge
```

The project is developed with assistance from AI coding tools. Every AI-assisted change must be reviewed and validated by a human contributor.

## Documentation

Planned documentation includes:

- `docs/ARCHITECTURE.md` for service boundaries and audio flow;
- `docs/DEPLOYMENT.md` for Docker Compose, TLS, DNS, firewall, TURN, and LiveKit;
- `docs/TROUBLESHOOTING.md` for common setup and WebRTC problems.

## License

See [LICENSE](LICENSE).
