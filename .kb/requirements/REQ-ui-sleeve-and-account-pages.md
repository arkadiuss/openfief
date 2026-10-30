---
title: The built-in UI must list sleeve kinds, sleeves, accounts with their connector instances, and broker accounts with the unmapped ones marked, for any session, and must give the owner forms that add and change kinds, sleeves and accounts and bind connectors, with a broker account picked from a list, handing each to the core's owner services and showing the core's reason when it refuses one.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: The built-in UI must list sleeve kinds, sleeves, accounts with their connector instances, and broker accounts with the unmapped ones marked, for any session, and must give the owner forms that add and change kinds, sleeves and accounts and bind connectors, with a broker account picked from a list, handing each to the core's owner services and showing the core's reason when it refuses one.
logic_claims:
  - CLAIM-BE4BA0029D3DB358
semantic_clauses:
  - The built-in UI must list sleeve kinds, sleeves, accounts with their connector instances, and broker accounts with the unmapped ones marked, for any session, and must give the owner forms that add and change kinds, sleeves and accounts and bind connectors, with a broker account picked from a list, handing each to the core's owner services and showing the core's reason when it refuses one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4e04f32bd35383e4e03c6124163b693394c35e3a9553ee0368c608d38052c393
semantic_inventory:
  - claim_key: CLAIM-BE4BA0029D3DB358
    claim_text: The built-in UI must list sleeve kinds, sleeves, accounts with their connector instances, and broker accounts with the unmapped ones marked, for any session, and must give the owner forms that add and change kinds, sleeves and accounts and bind connectors, with a broker account picked from a list, handing each to the core's owner services and showing the core's reason when it refuses one
    role: normative
    status: modeled
    span:
      start: 0
      end: 390
origin:
  kind: agent
  recorded_at: '2026-10-29T12:46:18.497Z'
id: REQ-ui-sleeve-and-account-pages
type: req
---
The built-in UI must list sleeve kinds, sleeves, accounts with their connector instances, and broker accounts with the unmapped ones marked, for any session, and must give the owner forms that add and change kinds, sleeves and accounts and bind connectors, with a broker account picked from a list, handing each to the core's owner services and showing the core's reason when it refuses one.

## Context

D-71 and D-72 have the owner manage sleeves and accounts in the built-in UI, outside the wizard too; the UI shows unmapped broker accounts (D-72). Refines REQ-ingest-owner-creates-accounts and REQ-ingest-broker-account-choice. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
