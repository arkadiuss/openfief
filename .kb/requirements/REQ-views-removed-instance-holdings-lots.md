---
title: Holdings and lots must be marked as no longer updated when any transaction behind them comes from a removed instance.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-91-1
  - lane:strict
text_ref: design-decisions:D-91#1
origin:
  kind: import
  ref: design-decisions:D-91#1
  recorded_at: '2026-10-29T08:36:32.879Z'
semantic_text: Holdings and lots must be marked as no longer updated when any transaction behind them comes from a removed instance.
logic_claims:
  - CLAIM-10339F77C271CE85
semantic_clauses:
  - Holdings and lots must be marked as no longer updated when any transaction behind them comes from a removed instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6008f0ed73d4ea5a3b6bd0a4fece886590c71ebb24eea00bc60ad052c5272e5b
semantic_inventory:
  - claim_key: CLAIM-10339F77C271CE85
    claim_text: Holdings and lots must be marked as no longer updated when any transaction behind them comes from a removed instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-views-removed-instance-holdings-lots
type: req
---
Holdings and lots must be marked as no longer updated when any transaction behind them comes from a removed instance.

## Context

D-58 keeps the data an uninstalled instance brought in, marked as coming from a removed instance (REQ-plugins-uninstall-keeps-data). REQ-views-removed-instance-mark, which the author has not reviewed, marks only the transactions the read API lists. D-91 extends the mark to the holdings and lots those transactions build, because they will no longer be updated either; REQ-views-removed-instance-mark-at-read records how the mark is worked out. The author recorded D-91 and D-92 after the design session that answered the implementation's questions.

## Source

> Holdings and lots are marked as no longer updated when any transaction behind them comes from a removed instance. Like the transactions' own mark, it is worked out when the data is read, so the log stays append-only.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-91#1
