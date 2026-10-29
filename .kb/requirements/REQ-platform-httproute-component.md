---
title: OpenFief must ship an optional kustomize component that adds a Gateway API HTTPRoute to the core's Service, for the owner's own Gateway and hostname.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-28-2
  - lane:strict
text_ref: design-decisions:D-28#2
origin:
  kind: import
  ref: design-decisions:D-28#2
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: OpenFief must ship an optional kustomize component that adds a Gateway API HTTPRoute to the core's Service, for the owner's own Gateway and hostname.
logic_claims:
  - CLAIM-C93B609FF70CB0B8
semantic_clauses:
  - OpenFief must ship an optional kustomize component that adds a Gateway API HTTPRoute to the core's Service, for the owner's own Gateway and hostname.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1b6ea38182ce28e32be799a095c5ea767fa053817e649ded34ece31f63acf652
semantic_inventory:
  - claim_key: CLAIM-C93B609FF70CB0B8
    claim_text: OpenFief must ship an optional kustomize component that adds a Gateway API HTTPRoute to the core's Service, for the owner's own Gateway and hostname
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
id: REQ-platform-httproute-component
type: req
---
OpenFief must ship an optional kustomize component that adds a Gateway API HTTPRoute to the core's Service, for the owner's own Gateway and hostname.

## Context

The author decided this on 2026-10-29 when reviewing the docs, which over-emphasized `kubectl port-forward` as the way into OpenFief. OpenFief still exposes nothing by default (REQ-platform-exposes-nothing-by-default): the component is optional, and the owner points the HTTPRoute at their own Gateway and hostname. The route reaches the core alone, since the core is the only entry point for browsers (D-81, REQ-core-only-browser-entry-point). Outside `localhost`, passkeys work only over HTTPS, so the route needs TLS. An Ingress or `kubectl port-forward` stays the owner's choice, and the docs explain what each exposes.

## Source

> OpenFief exposes nothing by default: every Service it creates is ClusterIP, the core's being the only way in (D-81), and it ships no tunnel. An optional kustomize component adds a Gateway API HTTPRoute to the core, which the owner points at their own Gateway and hostname. Any other way in, such as an Ingress or `kubectl port-forward`, is the owner's choice, and the docs explain what each exposes. Outside `localhost`, passkeys work only over HTTPS, so a route needs TLS.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-28#2
