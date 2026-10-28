---
title: The IBKR Flex Web Service client must fetch a statement with SendRequest and GetStatement, retry while IBKR reports the statement not ready up to an attempt limit, cap the statement's size, and send the token only as the documented query parameter to the configured endpoints, never in an error.
status: open
text_ref: docs/DESIGN.md#3.5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-3-5
  - lane:strict
semantic_text: The IBKR Flex Web Service client must fetch a statement with SendRequest and GetStatement, retry while IBKR reports the statement not ready up to an attempt limit, cap the statement's size, and send the token only as the documented query parameter to the configured endpoints, never in an error.
logic_claims:
  - CLAIM-0D4D370FBE89C811
semantic_clauses:
  - The IBKR Flex Web Service client must fetch a statement with SendRequest and GetStatement, retry while IBKR reports the statement not ready up to an attempt limit, cap the statement's size, and send the token only as the documented query parameter to the configured endpoints, never in an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 302b9e9e91bbfe2ee27bca8cf139cd3c754015d884292299f0469b8194fe1382
semantic_inventory:
  - claim_key: CLAIM-0D4D370FBE89C811
    claim_text: The IBKR Flex Web Service client must fetch a statement with SendRequest and GetStatement, retry while IBKR reports the statement not ready up to an attempt limit, cap the statement's size, and send the token only as the documented query parameter to the configured endpoints, never in an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 294
origin:
  kind: agent
  recorded_at: '2026-10-28T06:34:11.989Z'
id: REQ-ibkrflex-web-service-pull
type: req
---
The IBKR Flex Web Service client must fetch a statement with SendRequest and GetStatement, retry while IBKR reports the statement not ready up to an attempt limit, cap the statement's size, and send the token only as the documented query parameter to the configured endpoints, never in an error.

## Context

The M0 architecture's IBKR Flex connector does a daily pull through the Flex Web Service. DESIGN section 3.5 counts a leaked read-only credential such as the Flex token as a threat, since it still exposes the full account history, so the token never appears in errors and is never sent to the statement URL that IBKR's response names. Waiting is injected so tests need no sleeps. Refines REQ-AUTO-FD618A602F7F9106 and REQ-AUTO-41A1F4D687DAA771. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The protocol follows IBKR's public Flex Web Service (version 3) documentation; no IBKR credentials were available, so the replay recordings are synthetic.

## Source

> Data seen: every holding, transaction, statement file and account identifier, plus read-only broker credentials such as a Flex Query token.

docs/DESIGN.md section 3.5
