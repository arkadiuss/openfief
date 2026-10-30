---
title: An image CI does not rebuild on main must keep the digest the images workflow on main signed for an earlier commit, tagged for the new commit only after that signature verifies, so the demo deploy finds a signed digest for every image.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: An image CI does not rebuild on main must keep the digest the images workflow on main signed for an earlier commit, tagged for the new commit only after that signature verifies, so the demo deploy finds a signed digest for every image.
logic_claims:
  - CLAIM-EA5E2E006EC04D62
semantic_clauses:
  - An image CI does not rebuild on main must keep the digest the images workflow on main signed for an earlier commit, tagged for the new commit only after that signature verifies, so the demo deploy finds a signed digest for every image.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9b15123bb79ff712a31f505d88b034349b24619ab6ed7598c548f3e341b32512
semantic_inventory:
  - claim_key: CLAIM-EA5E2E006EC04D62
    claim_text: An image CI does not rebuild on main must keep the digest the images workflow on main signed for an earlier commit, tagged for the new commit only after that signature verifies, so the demo deploy finds a signed digest for every image
    role: normative
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-10-30T03:34:10.413Z'
id: REQ-tooling-ci-carry-signed-digest
type: req
---
An image CI does not rebuild on main must keep the digest the images workflow on main signed for an earlier commit, tagged for the new commit only after that signature verifies, so the demo deploy finds a signed digest for every image.

## Context

The demo deploy (D-17) pins every image at sha-<commit> to the digest whose keyless signature from images.yml on main verifies; an image left without that tag would fail the deploy, and one tagged on an unsigned digest must never be deployed. Tagging the signed digest without changing it keeps the signature valid, since a cosign signature belongs to the digest. Requested by the author through the orchestrating agent (modular, affected-only hook and CI); modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> CI builds, signs and publishes only the images whose binary's dependency closure, Dockerfile or shared build inputs changed; an unchanged image keeps the digest signed for an earlier commit, tagged for the new one once its signature verifies **(P)**, so the demo always deploys a signed digest. A pull request of docs or requirements alone skips the end-to-end and PostgreSQL jobs. Every push to `main` runs the whole suite, and a release builds every image.

docs/dev/sdlc.md#ci-and-environments
