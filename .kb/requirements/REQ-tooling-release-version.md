---
title: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable and YYYY.M.N-beta.K for beta with no v prefix, or whose tag already exists.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable and YYYY.M.N-beta.K for beta with no v prefix, or whose tag already exists.
logic_claims:
  - CLAIM-27B51BA2F272561D
semantic_clauses:
  - The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable and YYYY.M.N-beta.K for beta with no v prefix, or whose tag already exists.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f0fe07bccbef8832fad72647bf62e7d2ccc28d6d287d4c43432863f4a1336467
semantic_inventory:
  - claim_key: CLAIM-27B51BA2F272561D
    claim_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable and YYYY.M.N-beta.K for beta with no v prefix, or whose tag already exists
    role: normative
    status: modeled
    span:
      start: 0
      end: 179
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:45.838Z'
id: REQ-tooling-release-version
type: req
---
The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable and YYYY.M.N-beta.K for beta with no v prefix, or whose tag already exists.

## Context

Calendar versions and release channels are the author's release model in docs/dev/sdlc.md; the missing v prefix is a Claude Code proposal (P) because Go would read a root tag such as v2026.11.0 as major version 2026 of the module. Validating before anything is published keeps a typo from becoming a tag. Requested in the SDLC/CI implementation brief.

## Source

> Product versions are `YYYY.M.N`, with betas `YYYY.M.N-beta.K`. Tags carry no `v` prefix, because Go would read a root tag `v2026.11.0` as major version 2026 of the module **(P)**.

docs/dev/sdlc.md#releases
