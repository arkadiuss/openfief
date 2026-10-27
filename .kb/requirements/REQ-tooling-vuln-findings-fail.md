---
title: The vulnerability check must fail when govulncheck or npm audit reports a vulnerability.
status: open
text_ref: docs/dev/sdlc.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md
  - lane:strict
semantic_text: The vulnerability check must fail when govulncheck or npm audit reports a vulnerability.
logic_claims:
  - CLAIM-79AC155E56B1162F
semantic_clauses:
  - The vulnerability check must fail when govulncheck or npm audit reports a vulnerability.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f7810321cfbea48eb2475dd6743e5fa7f6bb1a7dbfb8d2db38782822a5072bfe
semantic_inventory:
  - claim_key: CLAIM-79AC155E56B1162F
    claim_text: The vulnerability check must fail when govulncheck or npm audit reports a vulnerability
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:34.802Z'
id: REQ-tooling-vuln-findings-fail
type: req
---
The vulnerability check must fail when govulncheck or npm audit reports a vulnerability.

## Context

Part of the local pipeline the author asked for: tests, lint and vulnerability check run on the pre-commit hook and in CI (D-36). Recorded by Claude Code while wiring the tooling under D-41.

## Source

> If a vulnerability database is unreachable, `task vuln` warns locally and fails in CI (`CI=true`). Findings always fail.

docs/dev/sdlc.md#commands
