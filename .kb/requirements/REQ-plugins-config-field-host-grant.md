---
title: Instead of a fixed host, a manifest may request the host of a named config field; the CLI must read that host from the instance's config and show it, and the grant must approve exactly that host.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-75-1
  - lane:strict
text_ref: design-decisions:D-75#1
origin:
  kind: import
  ref: design-decisions:D-75#1
  recorded_at: '2026-10-29T07:39:19.809Z'
semantic_text: Instead of a fixed host, a manifest may request the host of a named config field; the CLI must read that host from the instance's config and show it, and the grant must approve exactly that host.
logic_claims:
  - CLAIM-379112A106C0E870
semantic_clauses:
  - Instead of a fixed host, a manifest may request the host of a named config field; the CLI must read that host from the instance's config and show it, and the grant must approve exactly that host.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6d4cac65c7cd850be34a69fb862c73df7b3bc1c3d5d573e72cab9d020d1a4ce5
semantic_inventory:
  - claim_key: CLAIM-379112A106C0E870
    claim_text: Instead of a fixed host, a manifest may request the host of a named config field; the CLI must read that host from the instance's config and show it, and the grant must approve exactly that host
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
id: REQ-plugins-config-field-host-grant
type: req
---
Instead of a fixed host, a manifest may request the host of a named config field; the CLI must read that host from the instance's config and show it, and the grant must approve exactly that host.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames. D-74 and D-75 answer the implementation's question that the Ghostfolio view could not reach the owner's Ghostfolio: a manifest is the plugin's signed image label (D-42), so an image built and signed in CI can't name a host the owner picks at install, and REQ-plugins-ghostfolio-plugin records that its manifest names no egress host for that reason. D-75 lets the host come from the instance's own config (D-48); D-16 keeps what the manifest requests apart from what the owner's grant approves, and REQ-plugins-cli-permission-summary has the CLI show every requested egress host before the owner confirms. The author recorded D-75 with D-73 and D-74 after the design session that answered the implementation's questions.

## Source

> Instead of a fixed host, a manifest may request "the host of config field X". The CLI reads that host from the instance's config (D-48) and shows it, the grant approves exactly that host, and the plugin reaches it through the egress proxy like any other host. It must be a single HTTPS host: no wildcard and no IP address.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-75#1
