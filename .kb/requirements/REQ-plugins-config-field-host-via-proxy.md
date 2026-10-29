---
title: A plugin must reach a granted host taken from a config field through the egress proxy, like any other host.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-75-2
  - lane:strict
text_ref: design-decisions:D-75#2
origin:
  kind: import
  ref: design-decisions:D-75#2
  recorded_at: '2026-10-29T07:39:58.867Z'
semantic_text: A plugin must reach a granted host taken from a config field through the egress proxy, like any other host.
logic_claims:
  - CLAIM-D8FC7C76788E4AD0
semantic_clauses:
  - A plugin must reach a granted host taken from a config field through the egress proxy, like any other host.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d782e765862fa5cba74ff1e4ed4177b6e9c092a21cd4187237879e46112e52d7
semantic_inventory:
  - claim_key: CLAIM-D8FC7C76788E4AD0
    claim_text: A plugin must reach a granted host taken from a config field through the egress proxy, like any other host
    role: normative
    status: modeled
    span:
      start: 0
      end: 106
id: REQ-plugins-config-field-host-via-proxy
type: req
---
A plugin must reach a granted host taken from a config field through the egress proxy, like any other host.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames. D-75 lets a manifest request the host of a config field, which the grant approves exactly (REQ-plugins-config-field-host-grant). Unlike D-74's in-cluster service (REQ-plugins-in-cluster-service-no-proxy), such a host gets no exception from D-32: the egress proxy allows each plugin only the hostnames in its grant (REQ-AUTO-48C62B395F846700) and refuses private and internal address ranges (REQ-AUTO-79189DB32149E23F). The author recorded D-75 with D-73 and D-74 after the design session that answered the implementation's questions, among them that the Ghostfolio view could not reach the owner's Ghostfolio.

## Source

> Instead of a fixed host, a manifest may request "the host of config field X". The CLI reads that host from the instance's config (D-48) and shows it, the grant approves exactly that host, and the plugin reaches it through the egress proxy like any other host. It must be a single HTTPS host: no wildcard and no IP address.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-75#2
