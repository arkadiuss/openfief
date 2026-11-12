---
title: After a Ghostfolio sync, OpenFief-owned accounts must hold exactly the projected activities, a second sync must change nothing, a manually edited activity in an owned account must be reverted, and accounts OpenFief did not create must be untouched.
status: open
text_ref: design-decisions:D-200#9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-9
  - lane:strict
semantic_text: After a Ghostfolio sync, OpenFief-owned accounts must hold exactly the projected activities, a second sync must change nothing, a manually edited activity in an owned account must be reverted, and accounts OpenFief did not create must be untouched.
logic_claims:
  - CLAIM-27E72D2FE0617A6C
semantic_clauses:
  - After a Ghostfolio sync, OpenFief-owned accounts must hold exactly the projected activities, a second sync must change nothing, a manually edited activity in an owned account must be reverted, and accounts OpenFief did not create must be untouched.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e6700900a23baec536fbac1f22b8bbcd66e84293674353105eca9f9bc82db435
semantic_inventory:
  - claim_key: CLAIM-27E72D2FE0617A6C
    claim_text: After a Ghostfolio sync, OpenFief-owned accounts must hold exactly the projected activities, a second sync must change nothing, a manually edited activity in an owned account must be reverted, and accounts OpenFief did not create must be untouched
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#9
  recorded_at: '2026-11-08T08:59:34.210Z'
id: REQ-m0-exit-09-ghostfolio
type: req
---
After a Ghostfolio sync, OpenFief-owned accounts must hold exactly the projected activities, a second sync must change nothing, a manually edited activity in an owned account must be reverted, and accounts OpenFief did not create must be untouched.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-9.

## Source

> **Ghostfolio.** After a sync, Ghostfolio's OpenFief-owned accounts hold exactly the projected activities. A second sync changes nothing. A manually edited activity in an owned account is reverted, and accounts OpenFief did not create are untouched.

design-decisions:D-200#9 (docs/DESIGN.md, D-200, criterion 9)
