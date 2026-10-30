---
title: On a push to main, CI must build, sign and publish only the images whose Dockerfile, binary dependency closure, embedded files or shared build inputs changed since their last signed build, and a release must build every image.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: On a push to main, CI must build, sign and publish only the images whose Dockerfile, binary dependency closure, embedded files or shared build inputs changed since their last signed build, and a release must build every image.
logic_claims:
  - CLAIM-C6742DB6577042C4
semantic_clauses:
  - On a push to main, CI must build, sign and publish only the images whose Dockerfile, binary dependency closure, embedded files or shared build inputs changed since their last signed build, and a release must build every image.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2af88c7d496b98b28acdf1c15ee4ba3ed69794bdd9cff39a05b54ebab4ad6121
semantic_inventory:
  - claim_key: CLAIM-C6742DB6577042C4
    claim_text: On a push to main, CI must build, sign and publish only the images whose Dockerfile, binary dependency closure, embedded files or shared build inputs changed since their last signed build, and a release must build every image
    role: normative
    status: modeled
    span:
      start: 0
      end: 225
origin:
  kind: agent
  recorded_at: '2026-10-30T03:34:01.819Z'
id: REQ-tooling-ci-affected-images
type: req
---
On a push to main, CI must build, sign and publish only the images whose Dockerfile, binary dependency closure, embedded files or shared build inputs changed since their last signed build, and a release must build every image.

## Context

Building, signing and publishing all ten images on every push takes most of a CI run. An image's inputs are its Dockerfile, the packages its binary imports from go list -deps, the files they embed, and the module, workspace and workflow files every image shares. Comparing with the commit of the last signed build, not with the push's previous commit, keeps a cancelled or failed run from hiding a change. Releases build everything because they pin the catalog and tag every image with the version. Requested by the author through the orchestrating agent (modular, affected-only hook and CI); modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> CI builds, signs and publishes only the images whose binary's dependency closure, Dockerfile or shared build inputs changed; an unchanged image keeps the digest signed for an earlier commit, tagged for the new one once its signature verifies **(P)**, so the demo always deploys a signed digest. A pull request of docs or requirements alone skips the end-to-end and PostgreSQL jobs. Every push to `main` runs the whole suite, and a release builds every image.

docs/dev/sdlc.md#ci-and-environments

> The selector is a tested Go tool, and every hook and CI run prints what it selected and why.

docs/dev/sdlc.md#tooling
