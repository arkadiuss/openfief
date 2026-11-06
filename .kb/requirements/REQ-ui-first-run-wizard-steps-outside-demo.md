---
title: An owner session on an install that is not a demo and has no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.
status: open
text_ref: docs/DESIGN.md#D-191
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-191
  - lane:strict
semantic_text: An owner session on an install that is not a demo and has no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.
logic_claims:
  - CLAIM-78853C2E7A129AAD
semantic_clauses:
  - An owner session on an install that is not a demo and has no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 052085ea8d704ffebeec005e6779a059a675b81685c265efc753e2abbfb6e9bb
semantic_inventory:
  - claim_key: CLAIM-78853C2E7A129AAD
    claim_text: An owner session on an install that is not a demo and has no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step
    role: normative
    status: modeled
    span:
      start: 0
      end: 331
origin:
  kind: agent
  recorded_at: '2026-11-05T23:59:12.385Z'
id: REQ-ui-first-run-wizard-steps-outside-demo
type: req
---
An owner session on an install that is not a demo and has no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.

## Context

D-73: after the first passkey a wizard walks the owner through sleeves, accounts and binding connectors; installing plugins stays in the CLI. The UI knows the first run by the missing accounts, so the owner who created the first passkey lands on it; the steps and their order are the UI's reading. Refines REQ-ui-first-run-wizard-outside-demo. A demo has no wizard (D-191, REQ-ui-demo-no-wizard). Modeled by Claude Code; the author has not reviewed the wording.

## Why this changed

D-191 (author decision, 2026-11-05): a demo shows no first-run wizard. On a fresh demo database, before cmd/demo-seed wrote the demo marker, a visitor who enrolled a passkey landed in the wizard; the sleeves and accounts it stored were data without the marker, the loader's guard then refused the database as a real one, and the demo stayed stuck without View demo. So the wizard now belongs to installs that are not a demo; on a demo the front door says the demo data is loading until the marker exists (REQ-demo-no-wizard-until-loaded).

## Source

> **[D-73]** First-run setup happens in the built-in UI. After the owner creates the first passkey (D-60), a first-run wizard walks them through creating accounts (D-72), sleeves, and binding the connector instances already installed. Installing plugins stays in the CLI (D-46), whose last step points to the wizard.

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

docs/DESIGN.md#D-73, docs/DESIGN.md#D-191
