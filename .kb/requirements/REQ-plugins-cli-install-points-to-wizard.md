---
title: Installing plugins must stay in the openfief CLI, and the CLI's install must end by pointing the owner to the built-in UI's first-run wizard.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-73-2
  - lane:strict
text_ref: design-decisions:D-73#2
origin:
  kind: import
  ref: design-decisions:D-73#2
  recorded_at: '2026-10-29T07:34:47.187Z'
semantic_text: Installing plugins must stay in the openfief CLI, and the CLI's install must end by pointing the owner to the built-in UI's first-run wizard.
logic_claims:
  - CLAIM-C8614C56511E610C
semantic_clauses:
  - Installing plugins must stay in the openfief CLI, and the CLI's install must end by pointing the owner to the built-in UI's first-run wizard.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a5b2dbda63f9e8ae3e286ecece2a848d8507662a00f87d32bb774bb52e88d88a
semantic_inventory:
  - claim_key: CLAIM-C8614C56511E610C
    claim_text: Installing plugins must stay in the openfief CLI, and the CLI's install must end by pointing the owner to the built-in UI's first-run wizard
    role: normative
    status: modeled
    span:
      start: 0
      end: 140
id: REQ-plugins-cli-install-points-to-wizard
type: req
---
Installing plugins must stay in the openfief CLI, and the CLI's install must end by pointing the owner to the built-in UI's first-run wizard.

## Context

D-46 has plugins installed with the openfief CLI, which REQ-plugins-cli-install records. D-73 keeps installation there rather than in the built-in UI's first-run wizard, which binds the connector instances already installed (REQ-ui-first-run-wizard), and adds that the CLI's last step points the owner to that wizard. The author recorded D-73 with D-74 and D-75 after the design session that answered the implementation's questions; no further reason is stated.

## Source

> First-run setup happens in the built-in UI. After the owner creates the first passkey (D-60), a first-run wizard walks them through creating accounts (D-72), sleeves, and binding the connector instances already installed. Installing plugins stays in the CLI (D-46), whose last step points to the wizard.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-73#2
