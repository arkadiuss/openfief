---
title: 'The installed app must keep only the last overview on the device for offline reading: totals, each account''s values and the allocation''s shares, with the time they''re from.'
status: open
text_ref: design-decisions:D-151#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-151-1
  - lane:strict
semantic_text: 'The installed app must keep only the last overview on the device for offline reading: totals, each account''s values and the allocation''s shares, with the time they''re from.'
logic_claims:
  - CLAIM-7ABDF5913E9D3770
semantic_clauses:
  - 'The installed app must keep only the last overview on the device for offline reading: totals, each account''s values and the allocation''s shares, with the time they''re from.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 771ca651ab710f0c21a70e59e4475ead17ea792c1d66cca2dcd92f95b099ffc7
semantic_inventory:
  - claim_key: CLAIM-7ABDF5913E9D3770
    claim_text: 'The installed app must keep only the last overview on the device for offline reading: totals, each account''s values and the allocation''s shares, with the time they''re from'
    role: normative
    status: modeled
    span:
      start: 0
      end: 171
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-151#1
  recorded_at: '2026-11-01T12:28:29.896Z'
id: REQ-ui-offline-keeps-last-overview-only
type: req
---
The installed app must keep only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from.

## Context

The author recorded it on 2026-11-01. Reason not stated. It sits under the security controls for tracking and imports (GR-12), whose threats include theft of the data. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-151#1
