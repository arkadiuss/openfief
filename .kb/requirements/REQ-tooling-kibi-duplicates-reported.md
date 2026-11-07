---
title: When kibi's engine attaches a published store holding duplicate triples, it must keep one copy of each, and the next kibi command must report how many duplicate triples and entities stored twice it removed, as a workflow warning in CI.
status: open
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
semantic_text: When kibi's engine attaches a published store holding duplicate triples, it must keep one copy of each, and the next kibi command must report how many duplicate triples and entities stored twice it removed, as a workflow warning in CI.
logic_claims:
  - CLAIM-1B078F0E6E0509D3
semantic_clauses:
  - When kibi's engine attaches a published store holding duplicate triples, it must keep one copy of each, and the next kibi command must report how many duplicate triples and entities stored twice it removed, as a workflow warning in CI.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 42b1121968f653956f082b1034f550cb3131771fa63f739f62a85db1ebb6ac2f
semantic_inventory:
  - claim_key: CLAIM-1B078F0E6E0509D3
    claim_text: When kibi's engine attaches a published store holding duplicate triples, it must keep one copy of each, and the next kibi command must report how many duplicate triples and entities stored twice it removed, as a workflow warning in CI
    role: condition
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:04.442Z'
id: REQ-tooling-kibi-duplicates-reported
type: req
---
When kibi's engine attaches a published store holding duplicate triples, it must keep one copy of each, and the next kibi command must report how many duplicate triples and entities stored twice it removed, as a workflow warning in CI.

## Context

At 08788ac every kibi sync --rebuild published a store with 9,000 to 10,000 of its 17,468 entities stored twice, and duplicate answers multiplied in requirement proofs until the docs build's coverage query passed kibi's 120-second limit. The author approved a local kibi-core patch that removes the duplicates at attach time, on the condition that it stays loud so kibi's publish defect remains visible until it is fixed upstream. Requested by the author through the design session; modeled by Claude Code, not yet reviewed.

## Source

> KEEP IT LOUD: when the attach-time dedupe finds duplicates, print to stderr the number of duplicate triples and of duplicate entities removed.

docs/dev/kibi.md#setup-notes
