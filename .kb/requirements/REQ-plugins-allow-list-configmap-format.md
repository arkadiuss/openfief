---
title: The operator must hand egress allow-lists to the egress proxy as ConfigMap openfief-egress-allowlist with one key per plugin whose value lists canonical host:port lines, and the proxy must refuse data that is not in exactly that form.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The operator must hand egress allow-lists to the egress proxy as ConfigMap openfief-egress-allowlist with one key per plugin whose value lists canonical host:port lines, and the proxy must refuse data that is not in exactly that form.
logic_claims:
  - CLAIM-139EE916C1FD4853
semantic_clauses:
  - The operator must hand egress allow-lists to the egress proxy as ConfigMap openfief-egress-allowlist with one key per plugin whose value lists canonical host:port lines, and the proxy must refuse data that is not in exactly that form.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 50c701b7d21de97e168fe53b646907413b1855d31e921895278428dc28f9a0d8
semantic_inventory:
  - claim_key: CLAIM-139EE916C1FD4853
    claim_text: The operator must hand egress allow-lists to the egress proxy as ConfigMap openfief-egress-allowlist with one key per plugin whose value lists canonical host:port lines, and the proxy must refuse data that is not in exactly that form
    role: normative
    status: modeled
    span:
      start: 0
      end: 233
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:56.774Z'
id: REQ-plugins-allow-list-configmap-format
type: req
---
The operator must hand egress allow-lists to the egress proxy as ConfigMap openfief-egress-allowlist with one key per plugin whose value lists canonical host:port lines, and the proxy must refuse data that is not in exactly that form.

## Context

D-32 has the operator compile each grant's hostnames into a per-plugin allow-list that the proxy enforces; the operator and the proxy are separate processes, so they need one agreed format. One key per plugin keeps diffs readable and lets an empty value say a plugin has no egress. The decoder is strict, with the port always written, so that a typo can only narrow egress. Refines REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 egress proxy task (internal/plugins/app, smokescreen and kube adapters, cmd/egress-proxy); the author has not reviewed it yet.

## Source

> The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list.

docs/DESIGN.md D-32
