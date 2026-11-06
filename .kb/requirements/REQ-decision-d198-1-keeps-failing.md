---
title: A connector's sync must keep failing after three failed attempts in a row since its last success, spanning at least 15 minutes, and one success must clear it.
status: open
text_ref: design-decisions:D-198#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-198-1
  - lane:strict
semantic_text: A connector's sync must keep failing after three failed attempts in a row since its last success, spanning at least 15 minutes, and one success must clear it.
logic_claims:
  - CLAIM-FC88C5C98AF98965
semantic_clauses:
  - A connector's sync must keep failing after three failed attempts in a row since its last success, spanning at least 15 minutes, and one success must clear it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7bf1bc371413aaf7f4ec69a40db56de24b647f7befed0c0241ebfbc78b69c726
semantic_inventory:
  - claim_key: CLAIM-FC88C5C98AF98965
    claim_text: A connector's sync must keep failing after three failed attempts in a row since its last success, spanning at least 15 minutes, and one success must clear it
    role: normative
    status: modeled
    span:
      start: 0
      end: 157
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-198#1
  recorded_at: '2026-11-06T06:15:23.994Z'
id: REQ-decision-d198-1-keeps-failing
type: req
---
A connector's sync must keep failing after three failed attempts in a row since its last success, spanning at least 15 minutes, and one success must clear it.

## Context

Author decision D-198 (UI v2 design session) fixing D-196's 'keeps failing' as built.

## Source

> **[D-198]** A connector's sync keeps failing (D-196) after three failed attempts in a row since its last success, spanning at least 15 minutes; one success clears it. Every attempt is kept, with its error class and never its text (GR-12).

docs/DESIGN.md#D-198
