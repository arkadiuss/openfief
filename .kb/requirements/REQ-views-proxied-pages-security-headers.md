---
title: Every answer the core proxies from a view plugin, and the front door's own refusals, must carry the core's content security policy next to any the view sends, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: Every answer the core proxies from a view plugin, and the front door's own refusals, must carry the core's content security policy next to any the view sends, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer.
logic_claims:
  - CLAIM-CFCA1C09E8202DF1
semantic_clauses:
  - Every answer the core proxies from a view plugin, and the front door's own refusals, must carry the core's content security policy next to any the view sends, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e27d1df6ddcdc964f02151f9846b1f7d98ceb294a4d64c6ae0bd49001283596f
semantic_inventory:
  - claim_key: CLAIM-CFCA1C09E8202DF1
    claim_text: Every answer the core proxies from a view plugin, and the front door's own refusals, must carry the core's content security policy next to any the view sends, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer
    role: normative
    status: modeled
    span:
      start: 0
      end: 243
origin:
  kind: agent
  recorded_at: '2026-11-03T00:11:04.309Z'
id: REQ-views-proxied-pages-security-headers
type: req
---
Every answer the core proxies from a view plugin, and the front door's own refusals, must carry the core's content security policy next to any the view sends, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer.

## Context

A browser enforces every policy an answer carries, so the built-in UI's own stricter policy stays whole and a third-party view without one, or with a looser one, is held to the core's. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-184 existed in kibi on this branch when it was built.

## Source

> The database, the egress proxy and the operator each have a NetworkPolicy admitting only their callers. Pages the core proxies from a view plugin carry security headers the core sets: a content security policy, framing denied, no content sniffing and no referrer.

docs/DESIGN.md#D-184
