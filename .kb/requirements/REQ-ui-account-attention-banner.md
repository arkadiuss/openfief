---
title: An account's page must show a banner of its own open attention items at the top, each linked to the Inbox.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: An account's page must show a banner of its own open attention items at the top, each linked to the Inbox.
logic_claims:
  - CLAIM-A48DBCA949B20046
semantic_clauses:
  - An account's page must show a banner of its own open attention items at the top, each linked to the Inbox.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 93999788e09970609c16d921b4368d7416efb4439935d647bdb38c4e470d7745
semantic_inventory:
  - claim_key: CLAIM-A48DBCA949B20046
    claim_text: An account's page must show a banner of its own open attention items at the top, each linked to the Inbox
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
origin:
  kind: agent
  recorded_at: '2026-11-06T04:33:07.764Z'
id: REQ-ui-account-attention-banner
type: req
---
An account's page must show a banner of its own open attention items at the top, each linked to the Inbox.

## Context

D-196: an account's page shows its own items at the top; the gold banner follows the design system's M0 Account. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196
