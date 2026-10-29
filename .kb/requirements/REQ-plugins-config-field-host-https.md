---
title: A host taken from a config field must be a single HTTPS host, never a wildcard or an IP address.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-75-3
  - lane:strict
text_ref: design-decisions:D-75#3
origin:
  kind: import
  ref: design-decisions:D-75#3
  recorded_at: '2026-10-29T07:40:37.051Z'
semantic_text: A host taken from a config field must be a single HTTPS host, never a wildcard or an IP address.
logic_claims:
  - CLAIM-614A3F5A91D8271D
semantic_clauses:
  - A host taken from a config field must be a single HTTPS host, never a wildcard or an IP address.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7707e6d1b02a5edd86434cc7a735e263cfffe5dcf80ab0d8a16a95b2a7643c52
semantic_inventory:
  - claim_key: CLAIM-614A3F5A91D8271D
    claim_text: A host taken from a config field must be a single HTTPS host, never a wildcard or an IP address
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
id: REQ-plugins-config-field-host-https
type: req
---
A host taken from a config field must be a single HTTPS host, never a wildcard or an IP address.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames. D-75 lets a manifest request the host of a config field, which the grant approves exactly (REQ-plugins-config-field-host-grant) and the plugin reaches through the egress proxy (REQ-plugins-config-field-host-via-proxy); this requirement limits what that host may be. The implementation's REQ-plugins-egress-host-canonical already rejects IP literals and wildcards for every egress host in a manifest or grant. The author recorded D-75 with D-73 and D-74 after the design session that answered the implementation's questions; the decision states no further reason.

## Source

> Instead of a fixed host, a manifest may request "the host of config field X". The CLI reads that host from the instance's config (D-48) and shows it, the grant approves exactly that host, and the plugin reaches it through the egress proxy like any other host. It must be a single HTTPS host: no wildcard and no IP address.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-75#3
