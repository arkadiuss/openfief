---
title: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and its api-token Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
status: open
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
semantic_text: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and its api-token Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
logic_claims:
  - CLAIM-EA6ACC3BB2F37708
semantic_clauses:
  - The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and its api-token Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5b1330c7e5af8536e2fe92fb7083ef0145ef856ee3b0e0a749ed68f74bdbc948
semantic_inventory:
  - claim_key: CLAIM-EA6ACC3BB2F37708
    claim_text: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and its api-token Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice
    role: normative
    status: modeled
    span:
      start: 0
      end: 266
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:49.679Z'
id: REQ-plugins-eodhd-config
type: req
---
The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and its api-token Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.

## Context

D-51 and D-22 define the market-data connector; D-89 names its Secret by role, so each instance mounts <instance>-api-token. Refined by the plugin lifecycle package.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`.

docs/DESIGN.md#D-89
