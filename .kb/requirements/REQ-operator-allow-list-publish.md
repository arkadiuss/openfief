---
title: The operator must publish each running plugin's effective egress hosts under the plugin's key in the openfief-egress-allowlist ConfigMap and remove that key when the plugin stops running or is deleted.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The operator must publish each running plugin's effective egress hosts under the plugin's key in the openfief-egress-allowlist ConfigMap and remove that key when the plugin stops running or is deleted.
logic_claims:
  - CLAIM-FC6A5236BABD43AE
semantic_clauses:
  - The operator must publish each running plugin's effective egress hosts under the plugin's key in the openfief-egress-allowlist ConfigMap and remove that key when the plugin stops running or is deleted.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 648fbf4513db53141278379fbe8732bc8773d1988796ce1f1b3e44ee35ad4d4b
semantic_inventory:
  - claim_key: CLAIM-FC6A5236BABD43AE
    claim_text: The operator must publish each running plugin's effective egress hosts under the plugin's key in the openfief-egress-allowlist ConfigMap and remove that key when the plugin stops running or is deleted
    role: normative
    status: modeled
    span:
      start: 0
      end: 200
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:31.239Z'
id: REQ-operator-allow-list-publish
type: req
---
The operator must publish each running plugin's effective egress hosts under the plugin's key in the openfief-egress-allowlist ConfigMap and remove that key when the plugin stops running or is deleted.

## Context

D-32 has the operator compile each grant's hostnames into the plugin's allow-list for the egress proxy. The list holds effective hosts only, those both requested by the running version and granted. Each plugin's key has its own server-side apply field manager, so one reconcile never rewrites another plugin's entry from stale state. Refines REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list.

docs/DESIGN.md D-32
