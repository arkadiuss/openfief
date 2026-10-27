---
title: The OpenFief installation must not create an Ingress or a tunnel.
status: open
text_ref: design-decisions:D-28#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-28-1
  - lane:strict
semantic_text: The OpenFief installation must not create an Ingress or a tunnel.
logic_claims:
  - CLAIM-642D7736750EC51C
semantic_clauses:
  - The OpenFief installation must not create an Ingress or a tunnel.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 62219e996e117a1e4bc4999fa6ee92daebbcd75df9aefa4b5202543faceda83f
semantic_inventory:
  - claim_key: CLAIM-642D7736750EC51C
    claim_text: The OpenFief installation must not create an Ingress or a tunnel
    role: normative
    status: modeled
    span:
      start: 0
      end: 64
origin:
  kind: import
  ref: design-decisions:D-28#1
  recorded_at: '2026-10-27T20:12:45.783Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-3E7CF23DAD42DA83
type: req
---
The OpenFief installation must not create an Ingress or a tunnel.

## Source

> OpenFief ships no Ingress and no tunnel. The core and UI are ClusterIP services; the default way in is `kubectl port-forward`; anything else is the user's own Ingress, and the docs explain what it exposes.

Source: OpenFief author decisions (DESIGN.md section 10) - D-28#1
