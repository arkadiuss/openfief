---
title: The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything.
logic_claims:
  - CLAIM-E2F84D82D3E62780
semantic_clauses:
  - The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d0c1ab39555b9967db1616280d84d2cc3f977835ce363ba63a1cd8379185b710
semantic_inventory:
  - claim_key: CLAIM-E2F84D82D3E62780
    claim_text: The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything
    role: normative
    status: modeled
    span:
      start: 0
      end: 218
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:07.761Z'
id: REQ-operator-runs-only-granted
type: req
---
The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything.

## Context

D-16 splits what the vendor requests (Plugin) from what the owner approved (PluginGrant), and the operator enforces only what is in both. Naming the grant after the plugin gives each plugin at most one grant. Removing the ServiceAccount invalidates the plugin's tokens, and keeping a deny-all NetworkPolicy until the Plugin is deleted means a terminating pod reaches nothing. Refines REQ-AUTO-12177128A788CEDC. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> Plugin permissions are split into the `Plugin` resource (what the vendor requests) and a `PluginGrant` (what the user approved). The operator enforces only what is in both.

docs/DESIGN.md D-16
