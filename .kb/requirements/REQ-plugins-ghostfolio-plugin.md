---
title: The Ghostfolio plugin's manifest must ask only for the view capability, its security-token Secret, the host of its url config field and an in-cluster service; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach an external Ghostfolio only through the egress proxy and an in-cluster one only at the address the operator passes, and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
status: open
text_ref: docs/DESIGN.md#D-75
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-75
  - lane:strict
semantic_text: The Ghostfolio plugin's manifest must ask only for the view capability, its security-token Secret, the host of its url config field and an in-cluster service; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach an external Ghostfolio only through the egress proxy and an in-cluster one only at the address the operator passes, and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
logic_claims:
  - CLAIM-65B6293BE303885D
semantic_clauses:
  - The Ghostfolio plugin's manifest must ask only for the view capability, its security-token Secret, the host of its url config field and an in-cluster service; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach an external Ghostfolio only through the egress proxy and an in-cluster one only at the address the operator passes, and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d08a64bec84c3475def51764fc22e8ffdb16897d4aff65c786f7f9600e351d16
semantic_inventory:
  - claim_key: CLAIM-65B6293BE303885D
    claim_text: The Ghostfolio plugin's manifest must ask only for the view capability, its security-token Secret, the host of its url config field and an in-cluster service; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach an external Ghostfolio only through the egress proxy and an in-cluster one only at the address the operator passes, and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts
    role: normative
    status: modeled
    span:
      start: 0
      end: 627
origin:
  kind: agent
  recorded_at: '2026-10-29T01:37:01.053Z'
id: REQ-plugins-ghostfolio-plugin
type: req
---
The Ghostfolio plugin's manifest must ask only for the view capability, its security-token Secret, the host of its url config field and an in-cluster service; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach an external Ghostfolio only through the egress proxy and an in-cluster one only at the address the operator passes, and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.

## Context

D-3, D-4 and D-64 define the view; D-75 lets its manifest ask for the host of its url field, D-74 for a Ghostfolio in the same cluster, and D-89 names its Secret by role. Refined by the plugin lifecycle package.

## Source

> [D-75] Instead of a fixed host, a manifest may request "the host of config field X".

docs/DESIGN.md#D-75
