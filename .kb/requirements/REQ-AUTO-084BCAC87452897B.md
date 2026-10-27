---
title: The core and UI must be ClusterIP services.
status: open
text_ref: design-decisions:D-28#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-28-2
  - lane:strict
semantic_text: The core and UI must be ClusterIP services.
logic_claims:
  - CLAIM-0D54FAD338CCC932
semantic_clauses:
  - The core and UI must be ClusterIP services.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4736e8a1babac17f42b29d9de4c7735750a4d61441d99c336cf42cca269d461b
semantic_inventory:
  - claim_key: CLAIM-0D54FAD338CCC932
    claim_text: The core and UI must be ClusterIP services
    role: normative
    status: modeled
    span:
      start: 0
      end: 42
origin:
  kind: import
  ref: design-decisions:D-28#2
  recorded_at: '2026-10-27T20:12:46.445Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-084BCAC87452897B
type: req
---
The core and UI must be ClusterIP services.

## Source

> OpenFief ships no Ingress and no tunnel. The core and UI are ClusterIP services; the default way in is `kubectl port-forward`; anything else is the user's own Ingress, and the docs explain what it exposes.

Source: OpenFief author decisions (DESIGN.md section 10) - D-28#2
