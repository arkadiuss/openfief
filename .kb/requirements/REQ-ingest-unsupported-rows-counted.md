---
title: Each completed import batch must store how many of its rows OpenFief can't book yet, apart from rows left out on purpose.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: Each completed import batch must store how many of its rows OpenFief can't book yet, apart from rows left out on purpose.
logic_claims:
  - CLAIM-8DF65D4E28A65188
semantic_clauses:
  - Each completed import batch must store how many of its rows OpenFief can't book yet, apart from rows left out on purpose.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6f4f34353e8c7ef1f4fd7708324fb5eb674c019eaccdb8ef76dd9267bd103b64
semantic_inventory:
  - claim_key: CLAIM-8DF65D4E28A65188
    claim_text: Each completed import batch must store how many of its rows OpenFief can't book yet, apart from rows left out on purpose
    role: exception
    status: modeled
    span:
      start: 0
      end: 120
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:53.293Z'
id: REQ-ingest-unsupported-rows-counted
type: req
---
Each completed import batch must store how many of its rows OpenFief can't book yet, apart from rows left out on purpose.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). The import-problems check reads it.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
