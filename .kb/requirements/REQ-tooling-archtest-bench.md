---
title: Only the Ask box bench imports goeval, tests included, and only its packages that need nothing but the standard library; the bench alone, besides cmd, may import adapters and the demo dataset.
status: open
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:23:49.590Z'
firmness: default
semantic_text: Only the Ask box bench imports goeval, tests included, and only its packages that need nothing but the standard library; the bench alone, besides cmd, may import adapters and the demo dataset.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: adcd4e1d7b1f5a9eb0ae5a4552626dbd4480c4c639bfbc0093c1bc90c220b651
semantic_inventory:
  - claim_key: CLAIM-C3B906B3FD007F25
    claim_text: Only the Ask box bench imports goeval, tests included, and only its packages that need nothing but the standard library
    role: descriptive
    status: missing
    span:
      start: 0
      end: 119
  - claim_key: CLAIM-C3BE66E115781EC6
    claim_text: the bench alone, besides cmd, may import adapters and the demo dataset
    role: descriptive
    status: missing
    span:
      start: 121
      end: 191
logic_claims:
  - CLAIM-C3B906B3FD007F25
  - CLAIM-C3BE66E115781EC6
id: REQ-tooling-archtest-bench
type: req
---
Only the Ask box bench imports goeval, tests included, and only its packages that need nothing but the standard library; the bench alone, besides cmd, may import adapters and the demo dataset.

## Context

goeval's root, Gemini and embedding packages bring Google's client libraries into every build of the module; D-218 puts the harness in OpenFief and goeval behind it.

## Source

> The author's brief for the Ask box bench (AB1), 2026-11-14, on D-217, D-218 and D-219.
