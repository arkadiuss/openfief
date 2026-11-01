---
title: An owned Ghostfolio account for one currency of an account's cash must have a UUID derived from the view instance, the account and the currency.
status: open
text_ref: design-decisions:D-133#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-133-2
  - lane:strict
semantic_text: An owned Ghostfolio account for one currency of an account's cash must have a UUID derived from the view instance, the account and the currency.
logic_claims:
  - CLAIM-4E76005F052455F5
semantic_clauses:
  - An owned Ghostfolio account for one currency of an account's cash must have a UUID derived from the view instance, the account and the currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5ece2c308f891937603173e448debb111c6f9ad0965329337796986b0c2d7926
semantic_inventory:
  - claim_key: CLAIM-4E76005F052455F5
    claim_text: An owned Ghostfolio account for one currency of an account's cash must have a UUID derived from the view instance, the account and the currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 143
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-133#2
  recorded_at: '2026-11-01T07:39:08.651Z'
id: REQ-views-ghostfolio-cash-account-uuid
type: req
---
An owned Ghostfolio account for one currency of an account's cash must have a UUID derived from the view instance, the account and the currency.

## Context

The author recorded D-133 on 2026-10-31. D-64 has the Ghostfolio view derive its accounts' UUIDs so it knows which accounts it owns without storing anything: a plugin has no volume and no Kubernetes API access. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each currency of an account's cash other than its Ghostfolio account's gets an owned Ghostfolio account of its own, holding only that balance, with a UUID derived from the instance, the account and the currency, so Ghostfolio's totals include all of the account's cash.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-133#2
