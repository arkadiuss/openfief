---
title: The CLI must refuse to install over an existing instance.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-90-2
  - lane:strict
text_ref: design-decisions:D-90#2
origin:
  kind: import
  ref: design-decisions:D-90#2
  recorded_at: '2026-10-29T08:34:56.838Z'
semantic_text: The CLI must refuse to install over an existing instance.
logic_claims:
  - CLAIM-2B9925012487E9F8
semantic_clauses:
  - The CLI must refuse to install over an existing instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a3ce5e9772ddc268cd25433ea3de99aa44068c2fb20d2fa791cde7a7a8359a83
semantic_inventory:
  - claim_key: CLAIM-2B9925012487E9F8
    claim_text: The CLI must refuse to install over an existing instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 56
id: REQ-plugins-cli-no-install-over-instance
type: req
---
The CLI must refuse to install over an existing instance.

## Context

D-47 keeps plugin upgrades manual, and D-90 gives them their own command (REQ-plugins-cli-upgrade), so an install must never replace an instance that already exists. REQ-plugins-cli-install-verifies, which the author has not reviewed, already refuses a name that is already installed as part of a wider rule. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> The CLI refuses a catalog name that matches several versions and lists them instead of picking one, and refuses to install over an existing instance. `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-90#2
