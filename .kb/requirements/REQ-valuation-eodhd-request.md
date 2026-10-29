---
title: The EODHD adapter must send the owner's API key only as EODHD's api_token parameter, through the HTTP client and context its caller supplies, must never return the key in an error, must treat 404 as an unknown ticker and must fail on any other status than 200 or 404.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The EODHD adapter must send the owner's API key only as EODHD's api_token parameter, through the HTTP client and context its caller supplies, must never return the key in an error, must treat 404 as an unknown ticker and must fail on any other status than 200 or 404.
logic_claims:
  - CLAIM-8C920A6BB062F2BB
semantic_clauses:
  - The EODHD adapter must send the owner's API key only as EODHD's api_token parameter, through the HTTP client and context its caller supplies, must never return the key in an error, must treat 404 as an unknown ticker and must fail on any other status than 200 or 404.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4750ca4b11cba8c61a94e782485c190db6f4e00354d29ddb3ce21872ca458e76
semantic_inventory:
  - claim_key: CLAIM-8C920A6BB062F2BB
    claim_text: The EODHD adapter must send the owner's API key only as EODHD's api_token parameter, through the HTTP client and context its caller supplies, must never return the key in an error, must treat 404 as an unknown ticker and must fail on any other status than 200 or 404
    role: normative
    status: modeled
    span:
      start: 0
      end: 266
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:36.849Z'
id: REQ-valuation-eodhd-request
type: req
---
The EODHD adapter must send the owner's API key only as EODHD's api_token parameter, through the HTTP client and context its caller supplies, must never return the key in an error, must treat 404 as an unknown ticker and must fail on any other status than 200 or 404.

## Context

D-22 makes the owner bring the API key, and GR-12 forbids logging tokens; EODHD takes the key only in the URL, which Go's url.Error repeats, so the client strips it. The adapter runs in a plugin behind the egress proxy (D-32), so the caller owns the client. Refines REQ-AUTO-0C514DFCDB36FBEB. Modeled by Claude Code for the EODHD market-data connector task; the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22
