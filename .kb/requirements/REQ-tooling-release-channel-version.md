---
title: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists.
logic_claims:
  - CLAIM-4CFFF825D4F8EE6D
semantic_clauses:
  - The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cae54b3865d2315363a3fc0ecb19649c122d684426fb961b1188f3fce22e553c
semantic_inventory:
  - claim_key: CLAIM-4CFFF825D4F8EE6D
    claim_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists
    role: normative
    status: modeled
    span:
      start: 0
      end: 239
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:36.334Z'
id: REQ-tooling-release-channel-version
type: req
---
The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists.

## Context

docs/dev/sdlc.md (the author's release rules) adds the alpha channel, published exactly like a beta, and makes the first release an alpha. This supersedes REQ-tooling-release-version, which knew only beta and stable. The missing v prefix stays a Claude Code proposal (P). Modeled by Claude Code for release readiness item R3 (impl/release-alpha); the author has not reviewed it.

## Why this changed

The author's release rules in docs/dev/sdlc.md now add an alpha channel, published exactly like a beta, and make the first release an alpha, so the version check must accept YYYY.M.N-alpha.K where the superseded requirement knew only beta and stable versions.

## Source

> Product versions are `YYYY.M.N`, with pre-releases `YYYY.M.N-alpha.K` and `YYYY.M.N-beta.K`. Tags carry no `v` prefix, because Go would read a root tag `v2026.11.0` as major version 2026 of the module **(P)**.
>
> Channels: `alpha` and `beta` are pre-release channels, published the same way: a GitHub pre-release, deployed to demo. `stable` is a GitHub release. The first release is an alpha.

docs/dev/sdlc.md#releases
