---
title: River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once.
logic_claims:
  - CLAIM-BA71BFF910DC244F
semantic_clauses:
  - River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5ccfa74352a22bec67b3d34bea0272988ab66bd6a0337d6c908ec0758dbbb66b
semantic_inventory:
  - claim_key: CLAIM-BA71BFF910DC244F
    claim_text: River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once
    role: normative
    status: modeled
    span:
      start: 0
      end: 215
origin:
  kind: agent
  recorded_at: '2026-11-03T17:15:06.171Z'
id: REQ-valuation-river-migrations-serialized
type: req
---
River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once.

## Context

D-36 keeps the revaluation queue in River, whose migrator takes no lock of its own. The core and the demo loader both migrate on start and may start together (deploy/demo/openfief/kustomization.yaml). Found by CI run 41676052961 (e2e on k3d): the demo-seed Job and the core migrated River's schema at once and one failed on a duplicate key, and the Job spent its attempts while PostgreSQL was still starting. Modeled by Claude Code at the orchestrator's request; the author has not reviewed the wording.

## Source

> stored valuations ... River job queue

docs/DESIGN.md D-36
