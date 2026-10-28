---
title: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left.
status: open
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
semantic_text: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left.
logic_claims:
  - CLAIM-DB9445E3CB5F690B
semantic_clauses:
  - A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3473a342cfabb86343ab1c3e1fc19eb40d13d7054d111a13480bd87a7911b743
semantic_inventory:
  - claim_key: CLAIM-DB9445E3CB5F690B
    claim_text: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left
    role: normative
    status: modeled
    span:
      start: 0
      end: 284
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:10.232Z'
id: REQ-tooling-demo-bootstrap
type: req
---
A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left.

## Context

Pinning and checksums make a recreated machine identical and stop a tampered download; D-28 rules out ingress, and D-12 needs a CNI that enforces NetworkPolicy, which k3s's flannel with its policy controller provides. cloud-init user data is readable from the machine's metadata service, so it must hold no secret. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> OpenFief ships no Ingress and no tunnel.

docs/DESIGN.md D-28

> Tested on kind and k3d, with documented requirements: a CNI that enforces NetworkPolicy, and a storage class.

docs/DESIGN.md D-12
