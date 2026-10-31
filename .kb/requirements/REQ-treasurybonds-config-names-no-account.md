---
title: The treasury-bonds connector's config must name no account.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-118-3
  - lane:strict
text_ref: design-decisions:D-118#3
origin:
  kind: import
  ref: design-decisions:D-118#3
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The treasury-bonds connector's config must name no account.
logic_claims:
  - CLAIM-815F0783FA654065
semantic_clauses:
  - The treasury-bonds connector's config must name no account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 918d02622179740ae9721cc44aef573ca25a0fd282b0b29d7b56aabb7038c879
semantic_inventory:
  - claim_key: CLAIM-815F0783FA654065
    claim_text: The treasury-bonds connector's config must name no account
    role: normative
    status: modeled
    span:
      start: 0
      end: 58
id: REQ-treasurybonds-config-names-no-account
type: req
---
The treasury-bonds connector's config must name no account.

## Context

The treasury-bonds plugin's config named the OpenFief account it booked every lot into, so the owner typed an account ID at install, before the first-run wizard had created that account (question #4 on docs/implementation-questions.md); D-72 already spares the owner from typing broker account numbers. Since the core hands each lot to the connector with its account (REQ-ingest-bond-lot-handed-with-account), the author decided on 2026-10-31 that the config names none.

## Source

> The core hands each bond lot to the connector together with the account the owner entered it on, so one treasury-bonds instance serves every bond account and its config names none.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-118#3
