---
title: Until the first beta or stable release, proto:breaking must let a breaking contract change through only in a commit marked breaking; from the first beta only what an author decision names in the exceptions file.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: Until the first beta or stable release, proto:breaking must let a breaking contract change through only in a commit marked breaking; from the first beta only what an author decision names in the exceptions file.
logic_claims:
  - CLAIM-D2FC89A55D4E2375
semantic_clauses:
  - Until the first beta or stable release, proto:breaking must let a breaking contract change through only in a commit marked breaking; from the first beta only what an author decision names in the exceptions file.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 38ec10c8f00779d4e5e3deebdb035e50c4fc5e368c6308d31f1af9b843c8299a
semantic_inventory:
  - claim_key: CLAIM-D2FC89A55D4E2375
    claim_text: Until the first beta or stable release, proto:breaking must let a breaking contract change through only in a commit marked breaking; from the first beta only what an author decision names in the exceptions file
    role: normative
    status: modeled
    span:
      start: 0
      end: 210
origin:
  kind: agent
  recorded_at: '2026-11-01T20:54:41.066Z'
id: REQ-tooling-breaking-before-first-beta
type: req
---
Until the first beta or stable release, proto:breaking must let a breaking contract change through only in a commit marked breaking; from the first beta only what an author decision names in the exceptions file.

## Context

The author changed the release rules (design commit 109dc7e): until the first beta the read API and plugin contracts may break between releases, a breaking change marked breaking and listed in the release notes, with no author decision. The orchestrator asked proto:breaking to follow it, telling the stage from the tags, and never to switch the check off for unmarked commits.

## Source

> Until the first beta, the read API and the plugin contracts may break between releases: a breaking change is marked breaking and listed in the release notes, and needs no author decision.

docs/dev/sdlc.md#releases
