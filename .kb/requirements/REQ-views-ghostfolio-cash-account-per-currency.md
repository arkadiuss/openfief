---
title: Each currency of an account's cash other than its Ghostfolio account's must get an owned Ghostfolio account of its own, holding only that balance.
status: open
text_ref: design-decisions:D-133#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-133-1
  - lane:strict
semantic_text: Each currency of an account's cash other than its Ghostfolio account's must get an owned Ghostfolio account of its own, holding only that balance.
logic_claims:
  - CLAIM-9550AB64FDE3E45C
semantic_clauses:
  - Each currency of an account's cash other than its Ghostfolio account's must get an owned Ghostfolio account of its own, holding only that balance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 64634bbf42d00604f0fb55823c777a4a36a0d259bba1c98ac8b0241a746e4bca
semantic_inventory:
  - claim_key: CLAIM-9550AB64FDE3E45C
    claim_text: Each currency of an account's cash other than its Ghostfolio account's must get an owned Ghostfolio account of its own, holding only that balance
    role: normative
    status: modeled
    span:
      start: 0
      end: 145
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-133#1
  recorded_at: '2026-11-01T07:39:02.725Z'
id: REQ-views-ghostfolio-cash-account-per-currency
type: req
---
Each currency of an account's cash other than its Ghostfolio account's must get an owned Ghostfolio account of its own, holding only that balance.

## Context

The author recorded D-133 on 2026-10-31 for the Ghostfolio view plugin (D-64, D-108). The decision gives the reason: so Ghostfolio's totals include all of the account's cash, since a Ghostfolio account holds a balance in one currency. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each currency of an account's cash other than its Ghostfolio account's gets an owned Ghostfolio account of its own, holding only that balance, with a UUID derived from the instance, the account and the currency, so Ghostfolio's totals include all of the account's cash.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-133#1
