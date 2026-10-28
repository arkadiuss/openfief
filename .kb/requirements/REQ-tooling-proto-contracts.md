---
title: Changes to the protobuf contracts must pass buf lint with the STANDARD rules and buf breaking against the base revision, and the committed generated Go code must equal what the pinned generators produce from them.
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: Changes to the protobuf contracts must pass buf lint with the STANDARD rules and buf breaking against the base revision, and the committed generated Go code must equal what the pinned generators produce from them.
logic_claims:
  - CLAIM-314DA7F5493D86DD
semantic_clauses:
  - Changes to the protobuf contracts must pass buf lint with the STANDARD rules and buf breaking against the base revision, and the committed generated Go code must equal what the pinned generators produce from them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 837e9d886c7cb1fc60812977c8c6121f5de2195e6831a21be6e80677a5322d19
semantic_inventory:
  - claim_key: CLAIM-314DA7F5493D86DD
    claim_text: Changes to the protobuf contracts must pass buf lint with the STANDARD rules and buf breaking against the base revision, and the committed generated Go code must equal what the pinned generators produce from them
    role: normative
    status: modeled
    span:
      start: 0
      end: 212
origin:
  kind: agent
  recorded_at: '2026-10-28T07:37:32.493Z'
id: REQ-tooling-proto-contracts
type: req
---
Changes to the protobuf contracts must pass buf lint with the STANDARD rules and buf breaking against the base revision, and the committed generated Go code must equal what the pinned generators produce from them.

## Context

docs/dev/sdlc.md (proposal) says contracts and SDKs use SemVer and buf breaking gates every SDK release; running it in task check catches a break in the change that makes it. Tools are pinned in the Taskfile and run locally, with no remote plugins. The buf configuration covers every package under proto/openfief, the plugin contracts (D-11, D-49) and the read API (D-50) alike. Supports REQ-AUTO-BC327C2200FC820F and REQ-views-read-api-protobuf. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Contracts and SDKs use SemVer, independent of the product version. The Go SDK is tagged `sdk/go/vX.Y.Z`, as Go requires for a submodule **(P)**, and `buf breaking` gates every SDK release **(P)**.

docs/dev/sdlc.md
