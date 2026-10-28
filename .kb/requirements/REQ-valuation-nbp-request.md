---
title: The NBP adapter must request table A for at most 93 days per call, through the HTTP client and context its caller supplies and with its own User-Agent, and must fail on any status other than 200 or 404.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: The NBP adapter must request table A for at most 93 days per call, through the HTTP client and context its caller supplies and with its own User-Agent, and must fail on any status other than 200 or 404.
logic_claims:
  - CLAIM-9F4B9CC015776EA9
semantic_clauses:
  - The NBP adapter must request table A for at most 93 days per call, through the HTTP client and context its caller supplies and with its own User-Agent, and must fail on any status other than 200 or 404.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ed66526fdb3fef7b2b87bfe51530b7b38ea6f211e0a15e26dd0d4de5518cbb71
semantic_inventory:
  - claim_key: CLAIM-9F4B9CC015776EA9
    claim_text: The NBP adapter must request table A for at most 93 days per call, through the HTTP client and context its caller supplies and with its own User-Agent, and must fail on any status other than 200 or 404
    role: normative
    status: modeled
    span:
      start: 0
      end: 201
origin:
  kind: agent
  recorded_at: '2026-10-28T05:57:26.591Z'
id: REQ-valuation-nbp-request
type: req
---
The NBP adapter must request table A for at most 93 days per call, through the HTTP client and context its caller supplies and with its own User-Agent, and must fail on any status other than 200 or 404.

## Context

NBP's Web API serves at most 93 days per request; the use case splits longer ranges (REQ-valuation-fx-range-splitting). The adapter runs in a plugin behind HTTPS_PROXY (D-11, D-32), so the caller owns the client, timeout and proxy and the adapter reads no environment. api.nbp.pl answers 403 to Go's default User-Agent, so the adapter names itself. Refines REQ-AUTO-5769411DB9663682 and REQ-AUTO-D1568676DF81CAF7. Modeled by Claude Code for the M0 NBP FX source task; the author has not reviewed it yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20
