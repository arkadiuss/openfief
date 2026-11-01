---
title: The overview kept on the device must be encrypted with a key the browser can't export.
status: open
text_ref: design-decisions:D-151#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-151-2
  - lane:strict
semantic_text: The overview kept on the device must be encrypted with a key the browser can't export.
logic_claims:
  - CLAIM-51A51CE796D71FD0
semantic_clauses:
  - The overview kept on the device must be encrypted with a key the browser can't export.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 02ab3331ea740f95f3e851ad36d3c01f2a0aa172739a93878cccb818a101fe8a
semantic_inventory:
  - claim_key: CLAIM-51A51CE796D71FD0
    claim_text: The overview kept on the device must be encrypted with a key the browser can't export
    role: normative
    status: modeled
    span:
      start: 0
      end: 85
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-151#2
  recorded_at: '2026-11-01T12:28:40.905Z'
id: REQ-ui-offline-overview-unexportable-key
type: req
---
The overview kept on the device must be encrypted with a key the browser can't export.

## Context

The author recorded it on 2026-11-01. The overview is the only data the installed app keeps on the device (REQ-ui-offline-keeps-last-overview-only). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-151#2
