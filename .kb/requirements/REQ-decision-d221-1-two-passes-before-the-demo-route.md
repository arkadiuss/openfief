---
title: A model on a host joins D-207's demo route only after passing the Ask box's eval (D-220) in two consecutive runs, and leaves it after failing two consecutive weekly runs.
status: open
tags: []
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-221#1
  recorded_at: '2026-11-14T19:56:53.074Z'
text_ref: design-decisions:D-221#1
semantic_text: A model on a host joins D-207's demo route only after passing the Ask box's eval (D-220) in two consecutive runs, and leaves it after failing two consecutive weekly runs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bf3b18e30fb342539a9930a801566f04378bfb50cd221fab80201bbf9c17d050
semantic_inventory:
  - claim_key: CLAIM-A4E380478A693849
    claim_text: A model on a host joins D-207's demo route only after passing the Ask box's eval (D-220) in two consecutive runs, and leaves it after failing two consecutive weekly runs
    role: descriptive
    status: missing
    span:
      start: 0
      end: 169
logic_claims:
  - CLAIM-A4E380478A693849
id: REQ-decision-d221-1-two-passes-before-the-demo-route
type: req
---
A model on a host joins D-207's demo route only after passing the Ask box's eval (D-220) in two consecutive runs, and leaves it after failing two consecutive weekly runs.

## Context

Author decision D-221: how stable a pass must be before a model serves the demo, and what makes it leave. The bench keeps the history and gates the weekly workflow; the route itself is changed only in a reviewed change.

## Source

> **[D-221]** A model on a host joins D-207's demo route only after passing the Ask box's eval (D-220) in two consecutive runs, and leaves it after failing two consecutive weekly runs. A single failing run fails the weekly workflow, so the author sees it.
