---
title: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, through a Hetzner firewall with no other inbound rule, and the automation must refuse to create a server when no key has that label.
status: open
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
semantic_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, through a Hetzner firewall with no other inbound rule, and the automation must refuse to create a server when no key has that label.
logic_claims:
  - CLAIM-F44068081F69F05D
semantic_clauses:
  - A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, through a Hetzner firewall with no other inbound rule, and the automation must refuse to create a server when no key has that label.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 01268a6d6a43796ce3c37e43e538bc396691148a833a6ac9020a2fe3fb4193f6
semantic_inventory:
  - claim_key: CLAIM-F44068081F69F05D
    claim_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, through a Hetzner firewall with no other inbound rule, and the automation must refuse to create a server when no key has that label
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:42.816Z'
id: REQ-tooling-demo-access
type: req
---
A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, through a Hetzner firewall with no other inbound rule, and the automation must refuse to create a server when no key has that label.

## Context

D-28 says OpenFief ships no Ingress and the way in is kubectl port-forward, so the demo exposes nothing but SSH, which carries the port-forward tunnel; outbound stays open for Flux and image pulls. Key-only SSH means no password exists to leak. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> OpenFief ships no Ingress and no tunnel. The core and UI are ClusterIP services; the default way in is `kubectl port-forward`

docs/DESIGN.md D-28
