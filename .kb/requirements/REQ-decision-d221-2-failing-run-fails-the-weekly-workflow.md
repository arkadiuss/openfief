---
title: A single failing run fails the weekly workflow, so the author sees it.
status: open
tags: []
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-221#2
  recorded_at: '2026-11-14T19:56:55.029Z'
text_ref: design-decisions:D-221#2
semantic_text: A single failing run fails the weekly workflow, so the author sees it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1fadfdea65ab522f624358e5a462a4073716747dea2a21a7071358698995f5d9
semantic_inventory:
  - claim_key: CLAIM-01381850C366AD3C
    claim_text: A single failing run fails the weekly workflow, so the author sees it
    role: descriptive
    status: missing
    span:
      start: 0
      end: 69
logic_claims:
  - CLAIM-01381850C366AD3C
id: REQ-decision-d221-2-failing-run-fails-the-weekly-workflow
type: req
---
A single failing run fails the weekly workflow, so the author sees it.

## Context

Author decision D-221: how stable a pass must be before a model serves the demo, and what makes it leave. The bench keeps the history and gates the weekly workflow; the route itself is changed only in a reviewed change.

## Source

> **[D-221]** A model on a host joins D-207's demo route only after passing the Ask box's eval (D-220) in two consecutive runs, and leaves it after failing two consecutive weekly runs. A single failing run fails the weekly workflow, so the author sees it.
