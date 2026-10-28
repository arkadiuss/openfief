---
title: The egress proxy must follow the allow-list ConfigMap as it changes and, while it is missing or cannot be decoded, treat every plugin as having no egress allow-list.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The egress proxy must follow the allow-list ConfigMap as it changes and, while it is missing or cannot be decoded, treat every plugin as having no egress allow-list.
logic_claims:
  - CLAIM-FF247BD182B6C88B
semantic_clauses:
  - The egress proxy must follow the allow-list ConfigMap as it changes and, while it is missing or cannot be decoded, treat every plugin as having no egress allow-list.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8f6930f88c59216c502cbad22ac368cbab6397888a871a66529167e96c4aaecb
semantic_inventory:
  - claim_key: CLAIM-FF247BD182B6C88B
    claim_text: The egress proxy must follow the allow-list ConfigMap as it changes and, while it is missing or cannot be decoded, treat every plugin as having no egress allow-list
    role: normative
    status: modeled
    span:
      start: 0
      end: 164
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:39.604Z'
id: REQ-plugins-egress-allow-lists-fail-closed
type: req
---
The egress proxy must follow the allow-list ConfigMap as it changes and, while it is missing or cannot be decoded, treat every plugin as having no egress allow-list.

## Context

D-32 has the operator compile grants into allow-lists; grants change when the owner approves or revokes, so the proxy watches the ConfigMap rather than reading it once. Keeping the last good list after a bad update would keep revoked hosts reachable, so a bad or missing ConfigMap denies all egress and is reported for the operator to fix. Refines REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 egress proxy task (internal/plugins/app, smokescreen and kube adapters, cmd/egress-proxy); the author has not reviewed it yet.

## Source

> The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list.

docs/DESIGN.md D-32
