---
title: A config field a manifest names in egressFromConfig must give the host and port of the https URL it holds, lower-cased, nothing when it is absent or empty, and make the Plugin invalid when it is not an https URL of one hostname without a user part.
status: open
text_ref: docs/DESIGN.md#D-75
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-75
  - lane:strict
semantic_text: A config field a manifest names in egressFromConfig must give the host and port of the https URL it holds, lower-cased, nothing when it is absent or empty, and make the Plugin invalid when it is not an https URL of one hostname without a user part.
logic_claims:
  - CLAIM-C713554AE3C1C86B
semantic_clauses:
  - A config field a manifest names in egressFromConfig must give the host and port of the https URL it holds, lower-cased, nothing when it is absent or empty, and make the Plugin invalid when it is not an https URL of one hostname without a user part.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 65d62b5c2c092ab6fd527c452b3ddbdc68772ea6d230e86814fe8925a8b377f0
semantic_inventory:
  - claim_key: CLAIM-C713554AE3C1C86B
    claim_text: A config field a manifest names in egressFromConfig must give the host and port of the https URL it holds, lower-cased, nothing when it is absent or empty, and make the Plugin invalid when it is not an https URL of one hostname without a user part
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:55.075Z'
id: REQ-plugins-config-host-https-url
type: req
---
A config field a manifest names in egressFromConfig must give the host and port of the https URL it holds, lower-cased, nothing when it is absent or empty, and make the Plugin invalid when it is not an https URL of one hostname without a user part.

## Context

D-75 lets a manifest request the host of a config field and requires a single HTTPS host without wildcard or IP. Reading the field as an https URL and refusing a user part is the narrowest reading; an absent field requests nothing so a plugin such as Ghostfolio can be pointed at an in-cluster service instead (D-74). Refines REQ-plugins-config-field-host-https and REQ-plugins-config-field-host-grant.

## Source

> [D-75] Instead of a fixed host, a manifest may request "the host of config field X". The CLI reads that host from the instance's config (D-48) and shows it, the grant approves exactly that host, and the plugin reaches it through the egress proxy like any other host. It must be a single HTTPS host: no wildcard and no IP address.

docs/DESIGN.md#D-75
