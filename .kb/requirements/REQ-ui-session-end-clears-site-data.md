---
title: Sign-out and a session the core rejects must answer with Clear-Site-Data "storage" as the session cookie is dropped.
status: open
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
semantic_text: Sign-out and a session the core rejects must answer with Clear-Site-Data "storage" as the session cookie is dropped.
logic_claims:
  - CLAIM-307C0060A4410099
semantic_clauses:
  - Sign-out and a session the core rejects must answer with Clear-Site-Data "storage" as the session cookie is dropped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3a57e4c93c92e6898bacc5e5275246cb2a825966a3d21f13254cc197b614e329
semantic_inventory:
  - claim_key: CLAIM-307C0060A4410099
    claim_text: Sign-out and a session the core rejects must answer with Clear-Site-Data "storage" as the session cookie is dropped
    role: normative
    status: modeled
    span:
      start: 0
      end: 115
origin:
  kind: agent
  recorded_at: '2026-11-03T16:30:54.371Z'
id: REQ-ui-session-end-clears-site-data
type: req
---
Sign-out and a session the core rejects must answer with Clear-Site-Data "storage" as the session cookie is dropped.

## Context

D-151 wipes what the installed app keeps of the owner's record on sign-out or when the session ends. Clearing the origin's storage covers it whatever the app keeps, and the service worker with its cache. The review before the alpha tag (P1g) ported this part of D-151 ahead of the offline overview, which merges after the tag.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

docs/DESIGN.md#D-151
