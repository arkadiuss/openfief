---
title: By default, OpenFief must create no route, Ingress or tunnel to any of its Services.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-28-1
  - lane:strict
text_ref: design-decisions:D-28#1
origin:
  kind: import
  ref: design-decisions:D-28#1
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: By default, OpenFief must create no route, Ingress or tunnel to any of its Services.
logic_claims:
  - CLAIM-CBF13EEDC21CE96C
semantic_clauses:
  - By default, OpenFief must create no route, Ingress or tunnel to any of its Services.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5ba2ae23e474012d7cc97a5c63417facd9d90fdf81231f4ca1fb12f67d3ca97e
semantic_inventory:
  - claim_key: CLAIM-CBF13EEDC21CE96C
    claim_text: By default, OpenFief must create no route, Ingress or tunnel to any of its Services
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
id: REQ-platform-exposes-nothing-by-default
type: req
---
By default, OpenFief must create no route, Ingress or tunnel to any of its Services.

## Why this changed

REQ-AUTO-0D5A3B94C46E88DF made `kubectl port-forward` the default way in, and REQ-AUTO-3E7CF23DAD42DA83 kept the installation from creating an Ingress or a tunnel. In the docs session of 2026-10-29 the author changed D-28: `kubectl port-forward` is no longer the default way in but one of the owner's choices, next to their own Ingress, and OpenFief ships the core's ClusterIP Service plus an optional kustomize component with a Gateway API HTTPRoute (REQ-platform-httproute-component). What stays the default is stated here: OpenFief exposes nothing until the owner adds a way in.

## Context

The author decided this when reviewing the docs, which over-emphasized `kubectl port-forward` as the way into OpenFief. The core is the only way in for browsers (D-81, REQ-core-only-browser-entry-point), and the core's Service stays ClusterIP (REQ-AUTO-084BCAC87452897B). Outside `localhost`, passkeys work only over HTTPS, so any route the owner adds needs TLS.

## Source

> OpenFief exposes nothing by default: every Service it creates is ClusterIP, the core's being the only way in (D-81), and it ships no tunnel. An optional kustomize component adds a Gateway API HTTPRoute to the core, which the owner points at their own Gateway and hostname. Any other way in, such as an Ingress or `kubectl port-forward`, is the owner's choice, and the docs explain what each exposes. Outside `localhost`, passkeys work only over HTTPS, so a route needs TLS.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-28#1
