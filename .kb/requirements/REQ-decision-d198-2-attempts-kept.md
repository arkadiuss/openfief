---
title: Every sync attempt must be kept with its error class and never its error text.
status: open
text_ref: design-decisions:D-198#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-198-2
  - lane:strict
semantic_text: Every sync attempt must be kept with its error class and never its error text.
logic_claims:
  - CLAIM-9A3D07783EE061F4
semantic_clauses:
  - Every sync attempt must be kept with its error class and never its error text.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0210ad6505244d55882ae9185fb6caeafcda5922fa46efaccb928a92fe744e3e
semantic_inventory:
  - claim_key: CLAIM-9A3D07783EE061F4
    claim_text: Every sync attempt must be kept with its error class and never its error text
    role: normative
    status: modeled
    span:
      start: 0
      end: 77
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-198#2
  recorded_at: '2026-11-06T06:15:28.097Z'
id: REQ-decision-d198-2-attempts-kept
type: req
---
Every sync attempt must be kept with its error class and never its error text.

## Context

Author decision D-198; GR-12 keeps error text, which may carry paths or amounts, out of the record.

## Source

> **[D-198]** A connector's sync keeps failing (D-196) after three failed attempts in a row since its last success, spanning at least 15 minutes; one success clears it. Every attempt is kept, with its error class and never its text (GR-12).

docs/DESIGN.md#D-198
