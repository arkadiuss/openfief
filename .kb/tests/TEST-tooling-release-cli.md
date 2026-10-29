---
title: Every release builds the openfief CLI for Linux and macOS on amd64 and arm64 with an SBOM and checksums
status: active
sourceFile: scripts/tests/release-cli.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:56.181Z'
id: TEST-tooling-release-cli
type: test
---
Runs scripts/release-cli.sh with stand-ins for go and cyclonedx-gomod on PATH: exactly four builds, linux and darwin on amd64 and arm64, all with CGO disabled, named openfief-<version>-<os>-<arch>, an SBOM, and a SHA256SUMS that lists every file and verifies with sha256sum -c. Signing with cosign sign-blob runs only in CI.