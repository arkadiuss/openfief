---
title: The install guide must tell the owner to verify the CLI with cosign verify-blob against release.yml on main and the images with cosign verify against images.yml on main, both issued by GitHub Actions, and must name no other identity.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: The install guide must tell the owner to verify the CLI with cosign verify-blob against release.yml on main and the images with cosign verify against images.yml on main, both issued by GitHub Actions, and must name no other identity.
logic_claims:
  - CLAIM-A8D4DCE70CDC9D79
semantic_clauses:
  - The install guide must tell the owner to verify the CLI with cosign verify-blob against release.yml on main and the images with cosign verify against images.yml on main, both issued by GitHub Actions, and must name no other identity.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 19a9156f94d88d5c45c3734ad2408c9ebbebd8ed96f55f686db988b93f728bdc
semantic_inventory:
  - claim_key: CLAIM-A8D4DCE70CDC9D79
    claim_text: The install guide must tell the owner to verify the CLI with cosign verify-blob against release.yml on main and the images with cosign verify against images.yml on main, both issued by GitHub Actions, and must name no other identity
    role: normative
    status: modeled
    span:
      start: 0
      end: 232
origin:
  kind: agent
  recorded_at: '2026-11-01T14:59:03.905Z'
id: REQ-docs-install-guide-identities
type: req
---
The install guide must tell the owner to verify the CLI with cosign verify-blob against release.yml on main and the images with cosign verify against images.yml on main, both issued by GitHub Actions, and must name no other identity.

## Context

The install guide replaces the quickstart's note that an install comes with the first release. An identity in the guide that the release doesn't sign with would make every owner's verification fail, or worse, teach them to trust the wrong workflow, so a test ties the guide to the workflows. Requested by the orchestrator for release readiness item R4 (impl/release-alpha); the author has not reviewed it.

## Source

> Every release, an alpha included, ships signed images and the bootstrap install with the CRDs, every image in it pinned to the digest that was signed.
>
> Every release also ships the `openfief` CLI for Linux and macOS on amd64 and arm64, signed with cosign and with an SBOM (software bill of materials), since installing OpenFief starts with it (D-80).
>
> The release workflow is started manually, does a dry run before it publishes anything, and only the author releases.

docs/dev/sdlc.md#releases
