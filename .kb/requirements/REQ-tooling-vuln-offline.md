---
title: When a vulnerability database is unreachable, the vulnerability check must warn and pass locally, and must fail when CI is true.
status: closed
text_ref: docs/dev/sdlc.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md
  - lane:strict
semantic_text: When a vulnerability database is unreachable, the vulnerability check must warn and pass locally, and must fail when CI is true.
logic_claims:
  - CLAIM-1C02253D43CC2675
semantic_clauses:
  - When a vulnerability database is unreachable, the vulnerability check must warn and pass locally, and must fail when CI is true.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e3f8fa5cba4fadf14f6a181bd8c4a531bdb52db69674615f8f5b1800b03ee75e
semantic_inventory:
  - claim_key: CLAIM-1C02253D43CC2675
    claim_text: When a vulnerability database is unreachable, the vulnerability check must warn and pass locally, and must fail when CI is true
    role: condition
    status: modeled
    span:
      start: 0
      end: 127
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:49.668Z'
id: REQ-tooling-vuln-offline
type: req
---
When a vulnerability database is unreachable, the vulnerability check must warn and pass locally, and must fail when CI is true.

## Context

Keeps offline commits possible while CI stays strict. A Claude Code proposal, marked P in docs/dev/sdlc.md and awaiting the author's review.

## Source

> If a vulnerability database is unreachable, `task vuln` warns locally and fails in CI (`CI=true`). Findings always fail.

docs/dev/sdlc.md#tooling
