---
title: Before an install the CLI must show the instance name, the image, whether it is a verified plugin or only signed and by which signer, and every capability, egress host, Secret with its mount path, data tier and config property the manifest requests.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: Before an install the CLI must show the instance name, the image, whether it is a verified plugin or only signed and by which signer, and every capability, egress host, Secret with its mount path, data tier and config property the manifest requests.
logic_claims:
  - CLAIM-09914CC5B4702487
semantic_clauses:
  - Before an install the CLI must show the instance name, the image, whether it is a verified plugin or only signed and by which signer, and every capability, egress host, Secret with its mount path, data tier and config property the manifest requests.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 109a2a3e5e63a5aea4eeb8d90f7be45a5853a91fd4f9e528e7e85ad71ef8fc9c
semantic_inventory:
  - claim_key: CLAIM-09914CC5B4702487
    claim_text: Before an install the CLI must show the instance name, the image, whether it is a verified plugin or only signed and by which signer, and every capability, egress host, Secret with its mount path, data tier and config property the manifest requests
    role: normative
    status: modeled
    span:
      start: 0
      end: 248
origin:
  kind: agent
  recorded_at: '2026-10-28T23:36:15.036Z'
id: REQ-plugins-cli-permission-summary
type: req
---
Before an install the CLI must show the instance name, the image, whether it is a verified plugin or only signed and by which signer, and every capability, egress host, Secret with its mount path, data tier and config property the manifest requests.

## Context

D-46 has the CLI show the requested permissions before the owner confirms, and D-56 distinguishes a verified plugin from one only signed by a trusted signer. The grant approves exactly what is shown, so anything left out of the summary would be approved unseen. Secrets are listed with their mount paths because the owner creates them (D-29). Refines REQ-plugins-cli-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
