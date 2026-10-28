---
title: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and the eodhd-credentials Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and the eodhd-credentials Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
logic_claims:
  - CLAIM-3AF714C20F8C39A4
semantic_clauses:
  - The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and the eodhd-credentials Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f775fdf3abb34d5049993b2c71652f51c553bffca595028129cad67b99ec118c
semantic_inventory:
  - claim_key: CLAIM-3AF714C20F8C39A4
    claim_text: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and the eodhd-credentials Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice
    role: normative
    status: modeled
    span:
      start: 0
      end: 274
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:49.679Z'
id: REQ-plugins-eodhd-config
type: req
---
The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and the eodhd-credentials Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.

## Context

D-22: the owner brings the API key, so it comes from a Secret the owner creates (D-29) and the manifest mounts. The manifest is the signed image label (D-42), so it must ask for no more than the plugin needs. A config error found at start is better than a sync that silently prices nothing. Modeled by Claude Code for the EODHD market-data connector task; the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22
