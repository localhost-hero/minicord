# Contributing

Thank you for contributing to this project.

## AI-Assisted Development

This project is developed with assistance from AI coding tools, including GitHub Copilot. AI may help with exploration, implementation, tests, documentation, and code review, but it does not replace human ownership of the project.

Every change must be reviewed by a human contributor who is responsible for:

- understanding the proposed behavior and design;
- checking security, privacy, and licensing concerns;
- running the relevant tests and validation commands;
- verifying that the change follows `.github/copilot-instructions.md`;
- explaining AI-assisted changes when context is important for reviewers.

Do not submit unreviewed AI-generated code, fabricated test results, or claims about behavior that has not been verified.

The project is developed through small, reviewable changes. The default workflow is:

```text
Issue -> branch -> implementation -> tests -> commit -> Pull Request -> review -> merge
```

## Before You Start

1. Check existing issues and Pull Requests.
2. Create or select an Issue that describes the problem or feature.
3. Keep one logical change per branch and Pull Request.
4. Do not include secrets, personal infrastructure values, or unrelated formatting changes.

## Branches

Create a branch from the latest `main` branch:

```bash
git switch main
git pull --ff-only
git switch -c feat/short-description
```

Use English, lowercase names with one of these prefixes:

- `feat/` for a new feature
- `fix/` for a bug fix
- `docs/` for documentation
- `test/` for tests
- `refactor/` for code restructuring
- `chore/` for maintenance

Examples:

```text
feat/livekit-token-endpoint
fix/reconnect-call-state
docs/local-development
```

## Local Changes

Before editing, identify the module that owns the behavior. Keep changes focused and follow the architecture rules in `.github/copilot-instructions.md`.

For call-related changes, verify the complete lifecycle:

- room creation and access validation;
- participant connection;
- token issuance;
- audio routing through the SFU;
- mute and unmute;
- reconnecting and failure states;
- participant disconnect and room cleanup.

Never commit `.env` files, credentials, tokens, private keys, local paths, or home-network addresses.

## Validation

Run the narrowest relevant checks first. Run the following commands when the corresponding tooling exists in the repository:

```bash
go test ./...
go vet ./...
npm run lint
npm run build
```

For WebRTC and LiveKit changes, also verify that:

- audio is routed through the configured SFU or relay;
- direct peer-to-peer candidates are not used;
- unauthorized users cannot obtain room tokens;
- reconnect and permission-denied states remain understandable.

Document any check that could not be run and explain why.

## Commits

Write commit messages in English using this format:

```text
<type>: <short description>
```

Allowed types include:

- `feat`
- `fix`
- `docs`
- `test`
- `refactor`
- `chore`

Examples:

```text
feat: add room membership validation
fix: handle reconnecting call state
docs: describe local development setup
```

Keep commits small and logically complete. Do not use vague messages such as `changes`, `update`, or ` work`.

## Pull Requests

Push the branch and open a Pull Request against `main`:

```bash
git push -u origin feat/short-description
```

Use the Pull Request template. Describe:

- what changed;
- why it changed;
- how it was tested;
- configuration or migration requirements;
- known limitations or follow-up work.

A Pull Request should:

- have a clear link to its Issue;
- contain only related changes;
- pass applicable checks;
- update documentation when behavior, API, configuration, or deployment changes;
- avoid secrets and machine-specific values.

Do not merge until the change has been reviewed and required checks pass. If the project is maintained by one person, self-review the diff and test results before merging.

## Documentation

The primary language for code and documentation is English. Keep these documents updated when applicable:

- `README.md` for setup and project overview;
- `docs/ARCHITECTURE.md` for system boundaries and audio flow;
- `docs/DEPLOYMENT.md` for self-hosted deployment;
- `docs/TROUBLESHOOTING.md` for operational problems.
