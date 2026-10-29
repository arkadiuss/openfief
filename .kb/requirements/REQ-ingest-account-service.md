---
title: 'The core must serve the owner''s account entries as one service to view plugins acting for an owner session: creating an account with an ID made from its name, its institution, sleeve, the connector instance that feeds it and the broker account it is, binding the instance and mapping the broker account, changing an account''s name or sleeve, and binding an instance to an account, refusing a taken ID, an unknown sleeve or account, a malformed instance and a broker account the instance never reported or that is mapped to another account, before anything is stored.'
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: 'The core must serve the owner''s account entries as one service to view plugins acting for an owner session: creating an account with an ID made from its name, its institution, sleeve, the connector instance that feeds it and the broker account it is, binding the instance and mapping the broker account, changing an account''s name or sleeve, and binding an instance to an account, refusing a taken ID, an unknown sleeve or account, a malformed instance and a broker account the instance never reported or that is mapped to another account, before anything is stored.'
logic_claims:
  - CLAIM-005ADE0FE57798F1
semantic_clauses:
  - 'The core must serve the owner''s account entries as one service to view plugins acting for an owner session: creating an account with an ID made from its name, its institution, sleeve, the connector instance that feeds it and the broker account it is, binding the instance and mapping the broker account, changing an account''s name or sleeve, and binding an instance to an account, refusing a taken ID, an unknown sleeve or account, a malformed instance and a broker account the instance never reported or that is mapped to another account, before anything is stored.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f35c1e2d1877e801df518ca64fad61351636ca8af70941e19e2ad1cd05fb2621
semantic_inventory:
  - claim_key: CLAIM-005ADE0FE57798F1
    claim_text: 'The core must serve the owner''s account entries as one service to view plugins acting for an owner session: creating an account with an ID made from its name, its institution, sleeve, the connector instance that feeds it and the broker account it is, binding the instance and mapping the broker account, changing an account''s name or sleeve, and binding an instance to an account, refusing a taken ID, an unknown sleeve or account, a malformed instance and a broker account the instance never reported or that is mapped to another account, before anything is stored'
    role: normative
    status: modeled
    span:
      start: 0
      end: 565
origin:
  kind: agent
  recorded_at: '2026-10-29T12:46:04.649Z'
id: REQ-ingest-account-service
type: req
---
The core must serve the owner's account entries as one service to view plugins acting for an owner session: creating an account with an ID made from its name, its institution, sleeve, the connector instance that feeds it and the broker account it is, binding the instance and mapping the broker account, changing an account's name or sleeve, and binding an instance to an account, refusing a taken ID, an unknown sleeve or account, a malformed instance and a broker account the instance never reported or that is mapped to another account, before anything is stored.

## Context

D-72 has the owner create accounts in the UI with their institution, name, sleeve and connector instance, the broker account picked from those the connector reported; D-79 makes accounts their own owner-only service. Binding is what core connector bind does. Refines REQ-ingest-owner-creates-accounts and REQ-core-owner-entry-services. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72

> **[D-79]** What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69). The read API (D-6) and script tokens (D-27) stay read-only.

docs/DESIGN.md#D-79
