## Summary

<!-- What does this Pull Request change? -->

## Related Issue

<!-- Link the issue, for example: Closes #123 -->

## Testing

<!-- List the commands and manual checks that were run. -->

- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] Frontend lint/build, if configured
- [ ] Manual verification, if applicable

## Architecture and Security

- [ ] Audio still routes through the configured SFU or relay.
- [ ] No direct peer-to-peer fallback was added.
- [ ] Authorization is checked server-side.
- [ ] No secrets or personal infrastructure values were added.

## Documentation and Configuration

- [ ] Documentation was updated, or no update is required.
- [ ] Configuration and migration requirements are documented.
- [ ] `.env.example` was updated if variables changed.

## Notes

<!-- Known limitations, follow-up work, or deployment notes. -->
