---
title: Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release.
logic_claims:
  - CLAIM-4E9C525AC6B06AE7
semantic_clauses:
  - Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0f3403c615e318c816424ce4a8a3368efd864e7a1feeb8b43493a7d7a6bd3708
semantic_inventory:
  - claim_key: CLAIM-4E9C525AC6B06AE7
    claim_text: Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:17.775Z'
id: REQ-tooling-release-cli-signed
type: req
---
Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release.

## Context

docs/dev/sdlc.md (author rule) has every release ship the signed CLI with an SBOM, since installing OpenFief starts with it (D-80). A beta becomes a GitHub pre-release with only the CLI; that is this package's reading.

## Source

> Every release also ships the `openfief` CLI for Linux and macOS on amd64 and arm64, signed with cosign and with an SBOM (software bill of materials), since installing OpenFief starts with it (D-80).

docs/dev/sdlc.md#releases
