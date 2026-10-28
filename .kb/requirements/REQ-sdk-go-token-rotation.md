---
title: The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty.
logic_claims:
  - CLAIM-4779E08C64B66221
semantic_clauses:
  - The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 90b970a582e53ee1765b4b3646b3b9fcebb315eb7b6928f043e816881aba394c
semantic_inventory:
  - claim_key: CLAIM-4779E08C64B66221
    claim_text: The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty
    role: normative
    status: modeled
    span:
      start: 0
      end: 208
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:15.568Z'
id: REQ-sdk-go-token-rotation
type: req
---
The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty.

## Context

D-27 makes the projected ServiceAccount token (audience openfief-core) the plugin's credential towards the core, and D-63 the egress token towards the proxy. The kubelet rotates projected tokens before they expire, so a cached token would stop working. Refines REQ-AUTO-C6544ADAF8F5AE65. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`).

docs/DESIGN.md D-27
