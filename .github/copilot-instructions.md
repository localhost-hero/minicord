# Project AI Agent Instructions

## Project Purpose

This is a self-hosted web service for voice calls. The call architecture is `user -> server -> user`. Users must not connect directly to each other: the owner's server is a mandatory intermediary for call setup and audio transport.

## Architecture Rule

- Do not propose or implement a P2P route where audio can travel directly between clients.
- Do not treat a signaling server as sufficient for the `user-server-user` requirement: signaling exchanges metadata but does not transport media.
- For WebRTC, use forced server-side relaying through TURN (`iceTransportPolicy: "relay"`) or a server-side media relay/SFU. The selected technology must be documented and represented in configuration.
- Do not add a fallback to direct connections, STUN-only operation, or bypassing the owner's server without an explicit request from the project owner.
- All server addresses, ports, and relay modes must be provided through configuration and must not be hardcoded in the client.

## Planned Technology Stack

- Backend: Go. Use the standard library and small, proven dependencies; the Go server must control the application HTTP API, authentication, authorization, rooms, and LiveKit token issuance. LiveKit owns media-session signaling and media routing unless a separate requirement justifies an application-level WebSocket.
- Frontend: React + TypeScript + Vite. This provides a type-safe browser client, convenient call state management, and a small production bundle.
- Group voice calls: WebRTC + SFU. The preferred self-hosted option is LiveKit Server, which is written in Go and provides signaling, media routing, and client SDKs. Do not implement an SFU from scratch without a separate request.
- TURN: coturn as an additional relay for networks where a client cannot reach the SFU directly. When strict relay routing is required, keep `iceTransportPolicy: "relay"`.
- Audio codec: Opus.
- Deployment: Docker Compose; document the reverse proxy and TLS configuration.
- The database and authentication method are not fixed yet. Before adding either, choose the smallest option appropriate for the project scale and explain the trade-off.

## Group Call Plan

- The primary route is `browser -> SFU -> browsers`. A user sends one audio stream to the SFU, and the SFU forwards it to the required participants.
- The Go backend owns users, authorization, rooms, temporary token issuance, and SFU integration; do not proxy media traffic through ordinary Go HTTP handlers.
- For the first version, implement rooms, join/leave, mute/unmute, participant lists, active-speaker indication, and reliable reconnection.
- Limit the group size through server configuration and clean up empty rooms.
- Verify that a client receives a token only after the server validates access to the room.
- Integration tests must confirm that audio is routed through the SFU and that no direct peer-to-peer connection is used.

## Security and Privacy

- Use TLS by default for the web application, WebSocket, and WebRTC infrastructure; never disable certificate verification.
- Check authentication and authorization on the server for every action: joining a room, inviting a user, joining a call, and exchanging signaling messages.
- Do not trust user IDs, room IDs, ICE candidates, or call parameters received from a client.
- Do not log audio content, tokens, passwords, or private keys. Logs should contain only necessary technical events and anonymized identifiers.
- Do not add call recording by default. If recording is required, first document consent, storage, deletion, and access control.
- Store secrets only in environment variables or a secret manager. Never add them to source code, documentation, test fixtures, or Git.
- Account for home-server limits: rate limiting, message-size limits, timeouts, inactive-room cleanup, and graceful connection shutdown.

## Development Paradigm

- Use a pragmatic hybrid: apply object-oriented principles for encapsulation, contracts, and module responsibilities, while using idiomatic Go without imitating Java or C++ hierarchies.
- Use a modular monolith as the primary first-version architecture: one clear Go backend split into independent functional modules and one frontend.
- Organize code by functional slices rather than large global `controllers`, `services`, and `models` folders. Core slices are `auth`, `users`, `rooms`, `calls`, `participants`, `livekit`, and `health`.
- Within each module, separate transport/API, application use cases, domain rules, and infrastructure adapters only when this genuinely improves testing or dependency replacement.
- Use dependency inversion: Go business logic must not directly depend on LiveKit, a database, or a specific WebSocket package. Connect external systems through small interfaces and adapters.
- In Go, use structs with methods, small interfaces, composition, and dependency injection. Do not create classes, factories, or base types only to satisfy formal OOP.
- Prefer simple functions, explicit data structures, and composition over deep class hierarchies or universal abstractions.
- Model call state as an explicit state machine: `idle`, `connecting`, `connected`, `reconnecting`, `failed`, `ended`. Reject invalid transitions and test them.
- Use realtime events only for actual state changes: connect, disconnect, mute, unmute, and active-speaker changes. Do not turn every call into a hidden event chain.
- Use a component-based frontend with one-way state flow. Keep the WebRTC/LiveKit lifecycle in a dedicated layer or hook; visual components must not manage network connections directly.
- Do not add CQRS, event sourcing, microservices, or complex domain-driven design without proven need and a documented reason.
- Test module boundaries and keep them suitable for later extraction into a service, but do not split services prematurely.

## Change Behavior

- Before changing code, find the existing module that owns the behavior and follow the project's conventions.
- Make minimal, local changes; do not change the stack or public APIs without a clear need.
- For every call-related change, verify the full lifecycle: room creation, both users connecting, negotiation, audio through the relay, disconnection, reconnection, and resource cleanup.
- Handle microphone failure, permission denial, relay unavailability, network loss, reconnection, and call termination with clear user-facing states.
- Do not hide errors with fallback logic that violates `user-server-user`; show a diagnosable error and record a safe technical event.
- Add or update focused tests for server authorization, signaling routing, and relay rules. Integration tests must verify that direct ICE candidates are not used.
- After changes, run the narrowest relevant test, then linting/type checking and the build when available.
- Document startup commands, required environment variables, ports, TURN/relay settings, and reverse-proxy requirements.

## International Collaboration

- English is the primary language for source code, variable names, APIs, commit messages, comments, README files, documentation, and error messages. Do not add new Russian comments or entity names.
- Write documentation so a developer can clone or fork the project and run it from scratch without personal explanations from the author.
- Maintain an English `README.md` at the repository root with the project purpose, architecture, requirements, quick start, configuration, ports, and known limitations.
- Add `.env.example` with every required variable and no real secrets. Include a safe example and short description for each variable.
- Maintain `CONTRIBUTING.md` with installation, local development, testing, formatting, branch, and Pull Request rules.
- Maintain `docs/ARCHITECTURE.md` with the `browser -> Go backend -> LiveKit SFU -> browsers` route, service boundaries, and authorization/audio flow.
- Maintain `docs/DEPLOYMENT.md` with Docker Compose, reverse proxy, TLS, DNS, TURN/SFU, firewall ports, backups, and upgrades.
- Maintain `docs/TROUBLESHOOTING.md` with common problems: missing tokens, microphone permissions, unavailable SFU, blocked UDP ports, TLS failures, and WebRTC connection failures.
- Do not put secrets, personal data, author-local paths, or home IP addresses in documentation, fixtures, or configuration examples.
- State tool versions and verify documentation commands in a clean environment. Mark commands that depend on macOS, Linux, or Windows.
- When changing an API, configuration, data schema, or deployment, update the related documentation in the same change.
- Use neutral hostnames, domains, rooms, and user names for forks; do not bind code to the owner's personal infrastructure.

## Implementation Preferences

- First use existing dependencies and project patterns; add a new library only when there is a justified need.
- Keep the project lightweight in size, dependency count, memory use, CPU use, and deployment complexity.
- Additional size or complexity is allowed when it provides practical value, improves reliability, audio quality, performance, or security, or significantly reduces custom code.
- Before adding a heavy dependency or separate service, evaluate alternatives and briefly document the reason in the change or project documentation.
- Do not optimize prematurely: for voice media and group-call decisions, prioritize stability, low latency, and predictable behavior.
- Prefer one clear service over several small services when splitting provides no clear reliability, scaling, or security benefit.
- Keep server and client configuration separate, validate configuration at startup, and fail with a clear error when required parameters are missing.
- Keep the interface accessible with clear states for `connecting`, `call in progress`, `microphone permission denied`, `server unavailable`, and `call ended`.
- Use a highly minimalist interface without decorative blocks: the participant list, microphone state, leave control, and connection state must be available on the first screen.
- Prefer a calm neutral palette, clear typography, compact spacing, and predictable navigation. Do not add marketing hero sections, unnecessary cards, or visual noise.
- Use icons for mute, leave, and settings with accessible labels or tooltips; do not communicate critical states through color alone.
- Do not claim that audio travels through the server unless this is confirmed by relay configuration and verification of the actual route.
