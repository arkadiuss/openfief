---
title: openfief version must print the release version the release workflow stamped into the CLI, or dev for any other build, with the commit the binary was built from.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: openfief version must print the release version the release workflow stamped into the CLI, or dev for any other build, with the commit the binary was built from.
logic_claims:
  - CLAIM-A09F0466B8252B18
semantic_clauses:
  - openfief version must print the release version the release workflow stamped into the CLI, or dev for any other build, with the commit the binary was built from.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b38a7a0331768db7b70e8d6c4a61bb2d74a45c587d2cdaf1916c8261ae687eef
semantic_inventory:
  - claim_key: CLAIM-A09F0466B8252B18
    claim_text: openfief version must print the release version the release workflow stamped into the CLI, or dev for any other build, with the commit the binary was built from
    role: normative
    status: modeled
    span:
      start: 0
      end: 160
origin:
  kind: agent
  recorded_at: '2026-11-06T16:08:53.619Z'
id: REQ-tooling-cli-version
type: req
---
openfief version must print the release version the release workflow stamped into the CLI, or dev for any other build, with the commit the binary was built from.

## Context

Every release ships the openfief CLI (docs/dev/sdlc.md, D-80), and the owner downloads it from the GitHub release. The orchestrator asked for a version command after alpha.1 so an owner, and the post-release smoke test, can confirm a downloaded binary is the release it claims to be: scripts/release-cli.sh stamps the version with -ldflags -X main.version and builds with -buildvcs, so the commit comes from Go's own build information. A local build says dev.

## Source

> Every release also ships the `openfief` CLI for Linux and macOS on amd64 and arm64, signed with cosign and with an SBOM (software bill of materials), since installing OpenFief starts with it (D-80).

docs/dev/sdlc.md#releases
