---
title: 'First-run setup must happen in the built-in UI: after the owner creates the first passkey, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-73-1
  - lane:strict
text_ref: design-decisions:D-73#1
origin:
  kind: import
  ref: design-decisions:D-73#1
  recorded_at: '2026-10-29T07:34:08.320Z'
semantic_text: 'First-run setup must happen in the built-in UI: after the owner creates the first passkey, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.'
logic_claims:
  - CLAIM-03EC2A95AED92C27
semantic_clauses:
  - 'First-run setup must happen in the built-in UI: after the owner creates the first passkey, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 65586cfaa3fe6578df115f6f0976f762718c94bc3238e6d47f23fa95fa994b24
semantic_inventory:
  - claim_key: CLAIM-03EC2A95AED92C27
    claim_text: 'First-run setup must happen in the built-in UI: after the owner creates the first passkey, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed'
    role: normative
    status: modeled
    span:
      start: 0
      end: 216
id: REQ-ui-first-run-wizard
type: req
---
First-run setup must happen in the built-in UI: after the owner creates the first passkey, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.

## Context

D-72 answered the implementation's question that nothing could create an account or sleeve outside the demo, by having the owner create accounts in the built-in UI in an owner session (REQ-ingest-owner-creates-accounts). D-60 has the owner's first passkey created on localhost through a one-time link the CLI prints (REQ-identity-install-without-hostname). D-73 says where the first setup happens and in what order; installing plugins stays in the CLI, whose last step points to the wizard (REQ-plugins-cli-install-points-to-wizard). The author recorded D-73 with D-74 and D-75 after the design session that answered the implementation's questions; no further reason is stated.

## Source

> First-run setup happens in the built-in UI. After the owner creates the first passkey (D-60), a first-run wizard walks them through creating accounts (D-72), sleeves, and binding the connector instances already installed. Installing plugins stays in the CLI (D-46), whose last step points to the wizard.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-73#1
