---
title: The kept overview and its key must be wiped on sign-out (by the page before the form is sent, by the service worker and by a Clear-Site-Data "storage" response), when the core rejects the session (by the same header and on the login page), and offline, unread, once the session's end has passed.
status: open
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
semantic_text: The kept overview and its key must be wiped on sign-out (by the page before the form is sent, by the service worker and by a Clear-Site-Data "storage" response), when the core rejects the session (by the same header and on the login page), and offline, unread, once the session's end has passed.
logic_claims:
  - CLAIM-FB9BFF74A4D10FA4
semantic_clauses:
  - The kept overview and its key must be wiped on sign-out (by the page before the form is sent, by the service worker and by a Clear-Site-Data "storage" response), when the core rejects the session (by the same header and on the login page), and offline, unread, once the session's end has passed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 67ea8435b446c7aee5cad4bc49994f97c7446dc70c90db8f1dc49ec9f4601a77
semantic_inventory:
  - claim_key: CLAIM-FB9BFF74A4D10FA4
    claim_text: The kept overview and its key must be wiped on sign-out (by the page before the form is sent, by the service worker and by a Clear-Site-Data "storage" response), when the core rejects the session (by the same header and on the login page), and offline, unread, once the session's end has passed
    role: normative
    status: modeled
    span:
      start: 0
      end: 294
origin:
  kind: agent
  recorded_at: '2026-11-02T10:11:27.792Z'
id: REQ-ui-offline-overview-wipe-triggers
type: req
---
The kept overview and its key must be wiped on sign-out (by the page before the form is sent, by the service worker and by a Clear-Site-Data "storage" response), when the core rejects the session (by the same header and on the login page), and offline, unread, once the session's end has passed.

## Context

Refines REQ-ui-offline-overview-wiped-on-session-end. A session ends at the earlier of its idle and absolute expiry, or when revoked (D-66); the app can only learn it from the server or, offline, from the end the overview was rendered with. Recorded by the implementation agent of D-151 (impl/d151) with default firmness, refining the firm D-151 requirements the author recorded on 2026-11-01; the author has not reviewed this refinement yet.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

docs/DESIGN.md#D-151
