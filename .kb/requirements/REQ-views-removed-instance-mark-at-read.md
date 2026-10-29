---
title: The no-longer-updated mark on holdings and lots must be worked out when the data is read, so the log stays append-only.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-91-2
  - lane:strict
text_ref: design-decisions:D-91#2
origin:
  kind: import
  ref: design-decisions:D-91#2
  recorded_at: '2026-10-29T08:37:11.114Z'
semantic_text: The no-longer-updated mark on holdings and lots must be worked out when the data is read, so the log stays append-only.
logic_claims:
  - CLAIM-433C049536D1E06A
semantic_clauses:
  - The no-longer-updated mark on holdings and lots must be worked out when the data is read, so the log stays append-only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 3f55c448caf72f58c1e225d8bb7be9eff334958436c6c9b333904959e5591dff
semantic_inventory:
  - claim_key: CLAIM-433C049536D1E06A
    claim_text: The no-longer-updated mark on holdings and lots must be worked out when the data is read, so the log stays append-only
    role: normative
    status: modeled
    span:
      start: 0
      end: 118
id: REQ-views-removed-instance-mark-at-read
type: req
---
The no-longer-updated mark on holdings and lots must be worked out when the data is read, so the log stays append-only.

## Context

D-18 makes the transaction log append-only (REQ-AUTO-CACBF41A8E14837D), and REQ-views-removed-instance-mark, which the author has not reviewed, already works out the transactions' removed-instance mark when they are read. D-91 has the holdings-and-lots mark (REQ-views-removed-instance-holdings-lots) worked out the same way, at read time, so marking never writes to the log. The author recorded D-91 and D-92 after the design session that answered the implementation's questions.

## Source

> Holdings and lots are marked as no longer updated when any transaction behind them comes from a removed instance. Like the transactions' own mark, it is worked out when the data is read, so the log stays append-only.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-91#2
