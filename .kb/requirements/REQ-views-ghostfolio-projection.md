---
title: 'The Ghostfolio view must project each buy, sell and opening lot into a BUY or SELL at the unit price its holding posting implies with the fees in its currency, each dividend into a DIVIDEND of quantity 1 at the gross amount with the tax withheld as its fee, each fee into a FEE against the fee asset of its currency, fees or tax in another currency into FEEs of their own, and cash movements into nothing; must comment each activity with openfief: and the transaction''s key; must round every number to 15 significant digits; and must leave out, as a problem, a transaction whose listed instrument has no listing, whose numbers are too large or whose key is too long.'
status: open
text_ref: docs/DESIGN.md#D-4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-4
  - lane:strict
semantic_text: 'The Ghostfolio view must project each buy, sell and opening lot into a BUY or SELL at the unit price its holding posting implies with the fees in its currency, each dividend into a DIVIDEND of quantity 1 at the gross amount with the tax withheld as its fee, each fee into a FEE against the fee asset of its currency, fees or tax in another currency into FEEs of their own, and cash movements into nothing; must comment each activity with openfief: and the transaction''s key; must round every number to 15 significant digits; and must leave out, as a problem, a transaction whose listed instrument has no listing, whose numbers are too large or whose key is too long.'
logic_claims:
  - CLAIM-3D6C4600B62A27C8
semantic_clauses:
  - 'The Ghostfolio view must project each buy, sell and opening lot into a BUY or SELL at the unit price its holding posting implies with the fees in its currency, each dividend into a DIVIDEND of quantity 1 at the gross amount with the tax withheld as its fee, each fee into a FEE against the fee asset of its currency, fees or tax in another currency into FEEs of their own, and cash movements into nothing; must comment each activity with openfief: and the transaction''s key; must round every number to 15 significant digits; and must leave out, as a problem, a transaction whose listed instrument has no listing, whose numbers are too large or whose key is too long.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f4d6bd0890243d90607a51b1d19c5cfed5353333c4ec0d681aa1efef294a4893
semantic_inventory:
  - claim_key: CLAIM-3D6C4600B62A27C8
    claim_text: 'The Ghostfolio view must project each buy, sell and opening lot into a BUY or SELL at the unit price its holding posting implies with the fees in its currency, each dividend into a DIVIDEND of quantity 1 at the gross amount with the tax withheld as its fee, each fee into a FEE against the fee asset of its currency, fees or tax in another currency into FEEs of their own, and cash movements into nothing; must comment each activity with openfief: and the transaction''s key; must round every number to 15 significant digits; and must leave out, as a problem, a transaction whose listed instrument has no listing, whose numbers are too large or whose key is too long'
    role: normative
    status: modeled
    span:
      start: 0
      end: 665
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:48.245Z'
id: REQ-views-ghostfolio-projection
type: req
---
The Ghostfolio view must project each buy, sell and opening lot into a BUY or SELL at the unit price its holding posting implies with the fees in its currency, each dividend into a DIVIDEND of quantity 1 at the gross amount with the tax withheld as its fee, each fee into a FEE against the fee asset of its currency, fees or tax in another currency into FEEs of their own, and cash movements into nothing; must comment each activity with openfief: and the transaction's key; must round every number to 15 significant digits; and must leave out, as a problem, a transaction whose listed instrument has no listing, whose numbers are too large or whose key is too long.

## Context

D-4 puts OpenFief's transaction ID in each comment; the transaction key, source plus external ID, is that ID (D-18). Ghostfolio stores numbers as doubles, which print back any decimal of up to 15 significant digits unchanged, so rounding there keeps a second sync from seeing a difference. The mapping of kinds, quantity 1 for dividends and the fee assets are this view's choices; no decision fixes them. Refines REQ-AUTO-122F6AEE20C6FF6D. Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

## Source

> The Ghostfolio plugin runs a reconcile loop. It owns only the Ghostfolio accounts it created, puts OpenFief's transaction ID in each activity's comment, deletes stale activities and imports missing ones. Manual edits inside its accounts are reverted on the next sync.

docs/DESIGN.md D-4
