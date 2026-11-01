---
title: Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify.
logic_claims:
  - CLAIM-115A85ED364C01FA
semantic_clauses:
  - Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b324df8f11fbba8dece014e298a00b0d49d01f42e0d170b58cedbea55a1c24a9
semantic_inventory:
  - claim_key: CLAIM-115A85ED364C01FA
    claim_text: Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify
    role: normative
    status: modeled
    span:
      start: 0
      end: 250
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:48.902Z'
id: REQ-tooling-release-verify-signatures
type: req
---
Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify.

## Context

docs/dev/sdlc.md has every release ship signed images and a signed CLI and do a dry run before it publishes anything; the author wants the signing and the CLI exercised for real before the first alpha. Verifying with the same commands and identities the install guide gives the owner checks both the signatures and the guide. A dry run pushes no image, so it verifies the images main's last build signed. Modeled by Claude Code for release readiness items R3 and R4 (impl/release-alpha); the author has not reviewed it.

## Source

> Every release, an alpha included, ships signed images and the bootstrap install with the CRDs, every image in it pinned to the digest that was signed.
>
> Every release also ships the `openfief` CLI for Linux and macOS on amd64 and arm64, signed with cosign and with an SBOM (software bill of materials), since installing OpenFief starts with it (D-80).
>
> The release workflow is started manually, does a dry run before it publishes anything, and only the author releases.

docs/dev/sdlc.md#releases
