---
title: Default access to OpenFief must be through kubectl port-forward.
status: open
text_ref: design-decisions:D-28#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-28-3
  - lane:strict
semantic_text: Default access to OpenFief must be through kubectl port-forward.
logic_claims:
  - CLAIM-C301C258DBFAD81B
semantic_clauses:
  - Default access to OpenFief must be through kubectl port-forward.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0ebf3cfbc4a5aa0fd593e7622318d7fc1adb226531962dd57f555952f0a6e078
semantic_inventory:
  - claim_key: CLAIM-C301C258DBFAD81B
    claim_text: Default access to OpenFief must be through kubectl port-forward
    role: normative
    status: modeled
    span:
      start: 0
      end: 63
origin:
  kind: import
  ref: design-decisions:D-28#3
  recorded_at: '2026-10-27T20:12:47.169Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-0D5A3B94C46E88DF
type: req
---
Default access to OpenFief must be through kubectl port-forward.

## Source

> OpenFief ships no Ingress and no tunnel. The core and UI are ClusterIP services; the default way in is `kubectl port-forward`; anything else is the user's own Ingress, and the docs explain what it exposes.

Source: OpenFief author decisions (DESIGN.md section 10) - D-28#3
