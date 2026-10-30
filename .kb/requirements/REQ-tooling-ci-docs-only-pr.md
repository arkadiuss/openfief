---
title: CI must skip the end-to-end and PostgreSQL jobs for a pull request that changes only documentation or requirements, and must run them on every push to main.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: CI must skip the end-to-end and PostgreSQL jobs for a pull request that changes only documentation or requirements, and must run them on every push to main.
logic_claims:
  - CLAIM-73EAFF8824A5EF25
semantic_clauses:
  - CI must skip the end-to-end and PostgreSQL jobs for a pull request that changes only documentation or requirements, and must run them on every push to main.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8ae3d5a0d8df64c6e67d4ac74a78a88103c2580d8c1bd016644f2d3489c30b1e
semantic_inventory:
  - claim_key: CLAIM-73EAFF8824A5EF25
    claim_text: CI must skip the end-to-end and PostgreSQL jobs for a pull request that changes only documentation or requirements, and must run them on every push to main
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
origin:
  kind: agent
  recorded_at: '2026-10-30T03:34:21.502Z'
id: REQ-tooling-ci-docs-only-pr
type: req
---
CI must skip the end-to-end and PostgreSQL jobs for a pull request that changes only documentation or requirements, and must run them on every push to main.

## Context

The end-to-end matrix and the PostgreSQL tests cannot be affected by docs or requirements. Every push to main keeps the whole suite as the safety net for the selector, and task check stays whole on every pull request. Requested by the author through the orchestrating agent (modular, affected-only hook and CI); modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> CI builds, signs and publishes only the images whose binary's dependency closure, Dockerfile or shared build inputs changed; an unchanged image keeps the digest signed for an earlier commit, tagged for the new one once its signature verifies **(P)**, so the demo always deploys a signed digest. A pull request of docs or requirements alone skips the end-to-end and PostgreSQL jobs. Every push to `main` runs the whole suite, and a release builds every image.

docs/dev/sdlc.md#ci-and-environments
