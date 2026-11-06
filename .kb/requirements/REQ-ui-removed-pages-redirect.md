---
title: The paths of the removed check pages, /attention, /tickers, /assertions, /price-checks and /transfers, must redirect permanently to the Inbox.
status: open
text_ref: design-decisions:D-192
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-192
  - lane:strict
semantic_text: The paths of the removed check pages, /attention, /tickers, /assertions, /price-checks and /transfers, must redirect permanently to the Inbox.
logic_claims:
  - CLAIM-3A5D6522051FA549
semantic_clauses:
  - The paths of the removed check pages, /attention, /tickers, /assertions, /price-checks and /transfers, must redirect permanently to the Inbox.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 282bb683175a57b56aa45fa3329b0466ad75a1f0c8410f49fcb84e049b531f42
semantic_inventory:
  - claim_key: CLAIM-3A5D6522051FA549
    claim_text: The paths of the removed check pages, /attention, /tickers, /assertions, /price-checks and /transfers, must redirect permanently to the Inbox
    role: normative
    status: modeled
    span:
      start: 0
      end: 141
origin:
  kind: agent
  recorded_at: '2026-11-06T04:33:00.504Z'
id: REQ-ui-removed-pages-redirect
type: req
---
The paths of the removed check pages, /attention, /tickers, /assertions, /price-checks and /transfers, must redirect permanently to the Inbox.

## Context

The orchestrator asked that the old paths redirect to the Inbox; alpha may break URLs, but a bookmark should still land somewhere useful. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-192]** The built-in UI has four places: Overview, Accounts, Inbox and Settings. No check has a page of its own: what needs the owner gathers in the Inbox, and each check's results live on the account or instrument it is about.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-192
