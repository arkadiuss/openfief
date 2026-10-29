---
title: An owner session on an install with no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.
status: open
text_ref: docs/DESIGN.md#D-73
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-73
  - lane:strict
semantic_text: An owner session on an install with no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.
logic_claims:
  - CLAIM-0D170C93732582A6
semantic_clauses:
  - An owner session on an install with no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7db4cc7566ed4328c22d689bb7d8b88b6e55b0cedd52458c889a3ef645281d8a
semantic_inventory:
  - claim_key: CLAIM-0D170C93732582A6
    claim_text: An owner session on an install with no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step
    role: normative
    status: modeled
    span:
      start: 0
      end: 309
origin:
  kind: agent
  recorded_at: '2026-10-29T12:46:25.727Z'
id: REQ-ui-first-run-wizard-steps
type: req
---
An owner session on an install with no account must be sent from the overview to the first-run wizard, whose steps show the sleeves with their forms, then the accounts with the reported broker accounts and the installed connector instances to bind, then where to go next, with every form returning to its step.

## Context

D-73: after the first passkey a wizard walks the owner through sleeves, accounts and binding connectors; installing plugins stays in the CLI. The UI knows the first run by the missing accounts, so the owner who created the first passkey lands on it; the steps and their order are the UI's reading. Refines REQ-ui-first-run-wizard. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-73]** First-run setup happens in the built-in UI. After the owner creates the first passkey (D-60), a first-run wizard walks them through creating accounts (D-72), sleeves, and binding the connector instances already installed. Installing plugins stays in the CLI (D-46), whose last step points to the wizard.

docs/DESIGN.md#D-73
