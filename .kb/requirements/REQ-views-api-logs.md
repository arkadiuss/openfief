---
title: The read API must log requests only by an internal request ID, method, status and failing step, never a path, amount, holding, account identifier, token or error text, and must set no-store, nosniff, frame-denying, no-referrer and restrictive content-security headers on every response.
status: open
text_ref: docs/dev/code-organization.md#secrets-and-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-secrets-and-data
  - lane:strict
semantic_text: The read API must log requests only by an internal request ID, method, status and failing step, never a path, amount, holding, account identifier, token or error text, and must set no-store, nosniff, frame-denying, no-referrer and restrictive content-security headers on every response.
logic_claims:
  - CLAIM-7739A99185AEA8C0
semantic_clauses:
  - The read API must log requests only by an internal request ID, method, status and failing step, never a path, amount, holding, account identifier, token or error text, and must set no-store, nosniff, frame-denying, no-referrer and restrictive content-security headers on every response.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2228066db13654101a37c4a19eb048798c59cb044b836ade16ea28a35d2cdd09
semantic_inventory:
  - claim_key: CLAIM-7739A99185AEA8C0
    claim_text: The read API must log requests only by an internal request ID, method, status and failing step, never a path, amount, holding, account identifier, token or error text, and must set no-store, nosniff, frame-denying, no-referrer and restrictive content-security headers on every response
    role: normative
    status: modeled
    span:
      start: 0
      end: 285
origin:
  kind: agent
  recorded_at: '2026-10-28T08:44:31.119Z'
id: REQ-views-api-logs
type: req
---
The read API must log requests only by an internal request ID, method, status and failing step, never a path, amount, holding, account identifier, token or error text, and must set no-store, nosniff, frame-denying, no-referrer and restrictive content-security headers on every response.

## Context

The author's rule in docs/dev/code-organization.md (GR-12): never log amounts, holdings, account numbers or tokens; log internal IDs only. Paths name accounts and errors can quote amounts, so neither is logged; the problem's instance carries the request ID so the owner can match a failure to its log line. The headers keep browsers and proxies from caching or embedding financial JSON. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> Never log amounts, holdings, account numbers or tokens. Log internal IDs only (GR-12, UC-0).

docs/dev/code-organization.md (Secrets and data)
