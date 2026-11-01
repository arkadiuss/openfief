---
title: The overview kept on the device must be wiped on sign-out or when the session ends.
status: open
text_ref: design-decisions:D-151#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-151-3
  - lane:strict
semantic_text: The overview kept on the device must be wiped on sign-out or when the session ends.
logic_claims:
  - CLAIM-7C3BA20F2AB3F1F9
semantic_clauses:
  - The overview kept on the device must be wiped on sign-out or when the session ends.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1dd9ad9d19d018cebbcc6aa083a80d0151ecd7c024def9b0b983f4ca001392b4
semantic_inventory:
  - claim_key: CLAIM-7C3BA20F2AB3F1F9
    claim_text: The overview kept on the device must be wiped on sign-out or when the session ends
    role: normative
    status: modeled
    span:
      start: 0
      end: 82
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-151#3
  recorded_at: '2026-11-01T12:28:52.115Z'
id: REQ-ui-offline-overview-wiped-on-session-end
type: req
---
The overview kept on the device must be wiped on sign-out or when the session ends.

## Context

The author recorded it on 2026-11-01. Sessions expire after inactivity and an absolute lifetime and can be revoked (D-66). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-151#3
