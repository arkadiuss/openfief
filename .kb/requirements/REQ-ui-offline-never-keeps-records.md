---
title: The installed app must never keep transactions, lots or account numbers on the device.
status: open
text_ref: design-decisions:D-151#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-151-4
  - lane:strict
semantic_text: The installed app must never keep transactions, lots or account numbers on the device.
logic_claims:
  - CLAIM-168C1FBDA51BC064
semantic_clauses:
  - The installed app must never keep transactions, lots or account numbers on the device.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 58bee930545fe610f660aec3e5366199d6de300e1220d1cdd93775137d30eb25
semantic_inventory:
  - claim_key: CLAIM-168C1FBDA51BC064
    claim_text: The installed app must never keep transactions, lots or account numbers on the device
    role: normative
    status: modeled
    span:
      start: 0
      end: 85
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-151#4
  recorded_at: '2026-11-01T12:28:59.512Z'
id: REQ-ui-offline-never-keeps-records
type: req
---
The installed app must never keep transactions, lots or account numbers on the device.

## Context

The author recorded it on 2026-11-01. Only the overview is kept (REQ-ui-offline-keeps-last-overview-only); D-121 keeps broker account numbers to the owner's account service. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-151#4
