---
title: A balance assertion of cash in one currency or units of one instrument must be checked against the balance derived from the current versions at the end of its date, and a difference must be reported as a mismatch error that carries both the reported and the derived value.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: A balance assertion of cash in one currency or units of one instrument must be checked against the balance derived from the current versions at the end of its date, and a difference must be reported as a mismatch error that carries both the reported and the derived value.
logic_claims:
  - CLAIM-9B1FC97BD485DB0B
semantic_clauses:
  - A balance assertion of cash in one currency or units of one instrument must be checked against the balance derived from the current versions at the end of its date, and a difference must be reported as a mismatch error that carries both the reported and the derived value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6814d414828a0a6817cc6e6f4c1ef16635e2eb16f954c4d064b37b5cfa1d5972
semantic_inventory:
  - claim_key: CLAIM-9B1FC97BD485DB0B
    claim_text: A balance assertion of cash in one currency or units of one instrument must be checked against the balance derived from the current versions at the end of its date, and a difference must be reported as a mismatch error that carries both the reported and the derived value
    role: normative
    status: modeled
    span:
      start: 0
      end: 271
origin:
  kind: agent
  recorded_at: '2026-10-28T05:44:02.685Z'
id: REQ-ledger-assertion-mismatch
type: req
---
A balance assertion of cash in one currency or units of one instrument must be checked against the balance derived from the current versions at the end of its date, and a difference must be reported as a mismatch error that carries both the reported and the derived value.

## Context

Refines the firm requirement REQ-AUTO-D9FBE33256704D04 (D-18: broker position snapshots are stored as balance assertions, and a mismatch is a flagged error). Comparing at the end of the date matches how brokers report positions. Carrying both values lets the owner see how far off the ledger is without reading logs, which must not contain amounts (GR-12). Equality is numeric, so 1287 and 1287.00 match. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18

> A broker-reported balance on a date that the derived balance must match; a mismatch is an error.

docs/glossary.md (Balance assertion)
