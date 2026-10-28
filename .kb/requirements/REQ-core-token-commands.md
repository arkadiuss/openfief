---
title: core token issue must print a new token's secret once, as the first line on stdout, and log nothing about it; core token list must show every token's ID, name, scope, times and status and never a secret; core token revoke must revoke the token with the given ID.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: core token issue must print a new token's secret once, as the first line on stdout, and log nothing about it; core token list must show every token's ID, name, scope, times and status and never a secret; core token revoke must revoke the token with the given ID.
logic_claims:
  - CLAIM-3A844105AFEAAD85
semantic_clauses:
  - core token issue must print a new token's secret once, as the first line on stdout, and log nothing about it; core token list must show every token's ID, name, scope, times and status and never a secret; core token revoke must revoke the token with the given ID.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d4619e8e5d6440cbf70d95ff0ad4bd49071a10689aa354388d3ad17946c16419
semantic_inventory:
  - claim_key: CLAIM-3A844105AFEAAD85
    claim_text: core token issue must print a new token's secret once, as the first line on stdout, and log nothing about it; core token list must show every token's ID, name, scope, times and status and never a secret; core token revoke must revoke the token with the given ID
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:25.147Z'
id: REQ-core-token-commands
type: req
---
core token issue must print a new token's secret once, as the first line on stdout, and log nothing about it; core token list must show every token's ID, name, scope, times and status and never a secret; core token revoke must revoke the token with the given ID.

## Context

The owner gets tokens by running the binary in the pod (the image has no shell); printing the secret to stdout only keeps it out of logs (docs/dev/code-organization.md: never log tokens). Modeled by Claude Code for the core task: the author asked to see OpenFief's core running on the demo; the author has not reviewed this requirement yet.

## Source

> Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md D-27
