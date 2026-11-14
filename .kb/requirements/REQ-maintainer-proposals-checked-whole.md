---
title: A set of proposed target changes must pass every rule together, naming only known targets with exact decimal weights that sum to 100%, or be dropped whole and audited, never kept in part, with the note kept.
status: open
firmness: default
origin:
  kind: agent
  ref: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13
  recorded_at: '2026-11-13T21:07:47.341Z'
semantic_text: A set of proposed target changes must pass every rule together, naming only known targets with exact decimal weights that sum to 100%, or be dropped whole and audited, never kept in part, with the note kept.
semantic_source_field: semantic_text
semantic_source_hash: 715ad5b7f60bf776a33d8e20378dd9d0809b231829b5236f7f484a867626e6ee
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-9A59FC0512AD685B
    claim_text: A set of proposed target changes must pass every rule together, naming only known targets with exact decimal weights that sum to 100%, or be dropped whole and audited, never kept in part, with the note kept
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 206
logic_claims:
  - CLAIM-9A59FC0512AD685B
semantic_clauses:
  - A set of proposed target changes must pass every rule together, naming only known targets with exact decimal weights that sum to 100%, or be dropped whole and audited, never kept in part, with the note kept
required_evidence: unverified
id: REQ-maintainer-proposals-checked-whole
type: req
---
A set of proposed target changes must pass every rule together, naming only known targets with exact decimal weights that sum to 100%, or be dropped whole and audited, never kept in part, with the note kept.

## Context

The model is not trusted: its answer is checked before anything is stored, and an invalid proposal never reaches the Inbox.

## Source

> Validate proposals strictly (targets sum to 100% within the model's rules; known instruments only; decimal, never float64 for weights/money) - drop and audit invalid output.

Source: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13