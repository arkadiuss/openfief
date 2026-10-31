---
title: The UI must list attention items in one place.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-130-3
  - lane:strict
text_ref: design-decisions:D-130#3
origin:
  kind: import
  ref: design-decisions:D-130#3
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: The UI must list attention items in one place.
logic_claims:
  - CLAIM-100E125C62653185
semantic_clauses:
  - The UI must list attention items in one place.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 600ccca09ef504512d1614ab638349775d10157588ce23a859f23ca7bc93bd4b
semantic_inventory:
  - claim_key: CLAIM-100E125C62653185
    claim_text: The UI must list attention items in one place
    role: normative
    status: modeled
    span:
      start: 0
      end: 45
id: REQ-ui-attention-items-one-place
type: req
---
The UI must list attention items in one place.

## Context

The built-in UI shows what needs the owner on separate pages today: flagged balance assertions on /assertions (REQ-ui-account-pages), price checks on their own page with a count on the overview (REQ-ui-price-checks-page) and unmapped broker accounts on the accounts page, none of which the author has reviewed. D-130 puts every check that needs the owner into one store of attention items (REQ-core-attention-item-store), and the author asked that the UI list them in one place, as part of the one common mechanism for every flag and failed check they requested on 2026-10-31 while answering the implementation's questions.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks all use it, and the UI lists them in one place. Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-130#3
