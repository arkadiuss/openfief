---
title: The read API must list every sleeve kind with its default role, every connector binding, and the broker accounts connector instances reported with the account each is mapped to, the unmapped ones alone on request, and each sleeve's role override, as additions only.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: The read API must list every sleeve kind with its default role, every connector binding, and the broker accounts connector instances reported with the account each is mapped to, the unmapped ones alone on request, and each sleeve's role override, as additions only.
logic_claims:
  - CLAIM-6CE647D49411B134
semantic_clauses:
  - The read API must list every sleeve kind with its default role, every connector binding, and the broker accounts connector instances reported with the account each is mapped to, the unmapped ones alone on request, and each sleeve's role override, as additions only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8fc87cac265c27d8d8edd965a84e1bf36533a58ec255ae9a706bf0b01fee4857
semantic_inventory:
  - claim_key: CLAIM-6CE647D49411B134
    claim_text: The read API must list every sleeve kind with its default role, every connector binding, and the broker accounts connector instances reported with the account each is mapped to, the unmapped ones alone on request, and each sleeve's role override, as additions only
    role: normative
    status: modeled
    span:
      start: 0
      end: 264
origin:
  kind: agent
  recorded_at: '2026-10-29T12:46:11.268Z'
id: REQ-views-onboarding-lists
type: req
---
The read API must list every sleeve kind with its default role, every connector binding, and the broker accounts connector instances reported with the account each is mapped to, the unmapped ones alone on request, and each sleeve's role override, as additions only.

## Context

The UI shows kinds (D-71), which connector feeds each account and the unmapped broker accounts (D-72) through the read API (D-6). Broker account IDs are account numbers; any read caller may list them, as it may list the record itself. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72

> **[D-71]** Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

docs/DESIGN.md#D-71
