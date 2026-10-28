---
title: The egress proxy must resolve an allowed hostname itself and refuse the request when any resolved address fails the forbidden-address check, dialing only checked addresses.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The egress proxy must resolve an allowed hostname itself and refuse the request when any resolved address fails the forbidden-address check, dialing only checked addresses.
logic_claims:
  - CLAIM-D99BF274B0E8BB3F
semantic_clauses:
  - The egress proxy must resolve an allowed hostname itself and refuse the request when any resolved address fails the forbidden-address check, dialing only checked addresses.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1f70e248a06681cb31eeb24a85b6dd8d113673a2c89e684e246ea22ebf7f82ec
semantic_inventory:
  - claim_key: CLAIM-D99BF274B0E8BB3F
    claim_text: The egress proxy must resolve an allowed hostname itself and refuse the request when any resolved address fails the forbidden-address check, dialing only checked addresses
    role: normative
    status: modeled
    span:
      start: 0
      end: 171
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:32.725Z'
id: REQ-plugins-egress-proxy-checks-every-address
type: req
---
The egress proxy must resolve an allowed hostname itself and refuse the request when any resolved address fails the forbidden-address check, dialing only checked addresses.

## Context

D-32 has the proxy refuse private and internal ranges. Smokescreen alone would skip a forbidden address and dial a public one from the same answer; refusing the whole name instead keeps a hostname that mixes public and internal records from being usable at all and is simpler to reason about. Smokescreen's own deny ranges are set to the same list as a second layer. Refines REQ-AUTO-79189DB32149E23F. Modeled by Claude Code for the M0 egress proxy task (internal/plugins/app, smokescreen and kube adapters, cmd/egress-proxy); the author has not reviewed it yet.

## Source

> the proxy refuses private and internal address ranges

docs/DESIGN.md D-32
