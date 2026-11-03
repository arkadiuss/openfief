---
title: Grants made before D-185 must approve the institutions their plugin already declares.
status: open
text_ref: design-decisions:D-185#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-185-3
  - lane:strict
semantic_text: Grants made before D-185 must approve the institutions their plugin already declares.
logic_claims:
  - CLAIM-0B3C9CAB200F0D7A
semantic_clauses:
  - Grants made before D-185 must approve the institutions their plugin already declares.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b69a88cae63df71d5d65d0ade70ad3e7b4f020902dedbc6b171d7b3191988c22
semantic_inventory:
  - claim_key: CLAIM-0B3C9CAB200F0D7A
    claim_text: Grants made before D-185 must approve the institutions their plugin already declares
    role: normative
    status: modeled
    span:
      start: 0
      end: 84
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-185#3
  recorded_at: '2026-11-03T07:22:57.463Z'
id: REQ-plugins-legacy-grants-approve-declared-institutions
type: req
---
Grants made before D-185 must approve the institutions their plugin already declares.

## Context

The author decided it on 2026-11-03, before the alpha tag, extending D-182's declared institutions: what a connector books for becomes something the owner approves, not only something the vendor declares. So existing installs, the demo's Flux-installed plugins among them, keep syncing after the upgrade. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The institutions a connector books for are a permission its grant approves, as its hostnames are (D-16): the signed manifest declares them and the core accepts only what both name. The install warns when another instance already books for an institution, naming that instance. Grants made before this approve the institutions their plugin already declares.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-185#3, decided by the author on 2026-11-03
