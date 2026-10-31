---
title: The core must hand each bond lot to the connector together with the account the owner entered it on.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-118-1
  - lane:strict
text_ref: design-decisions:D-118#1
origin:
  kind: import
  ref: design-decisions:D-118#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The core must hand each bond lot to the connector together with the account the owner entered it on.
logic_claims:
  - CLAIM-A5F9BDE07DEB6A06
semantic_clauses:
  - The core must hand each bond lot to the connector together with the account the owner entered it on.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e31e0605b2b917b168e6353e7fb2b0956fe10b64c10288c425e1750938340a58
semantic_inventory:
  - claim_key: CLAIM-A5F9BDE07DEB6A06
    claim_text: The core must hand each bond lot to the connector together with the account the owner entered it on
    role: normative
    status: modeled
    span:
      start: 0
      end: 99
id: REQ-ingest-bond-lot-handed-with-account
type: req
---
The core must hand each bond lot to the connector together with the account the owner entered it on.

## Context

The treasury-bonds plugin booked every lot into the account its config's `account` named, so the owner had to type an OpenFief account ID at install, before the first-run wizard had created the account, unlike D-72's broker accounts, which the owner picks rather than types (question #4 on docs/implementation-questions.md). The core already knows the account each lot is entered on, so on 2026-10-31 the author decided it sends that account with the lot; one instance then serves every bond account (REQ-treasurybonds-one-instance-all-accounts) and its config names none (REQ-treasurybonds-config-names-no-account).

## Source

> The core hands each bond lot to the connector together with the account the owner entered it on, so one treasury-bonds instance serves every bond account and its config names none.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-118#1
