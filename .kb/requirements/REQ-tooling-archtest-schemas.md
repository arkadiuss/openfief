---
title: The architecture tests must fail when a context's SQL, in its migrations or in its packages' query strings, names a schema that another context owns.
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: The architecture tests must fail when a context's SQL, in its migrations or in its packages' query strings, names a schema that another context owns.
logic_claims:
  - CLAIM-14088189B2AC0C8B
semantic_clauses:
  - The architecture tests must fail when a context's SQL, in its migrations or in its packages' query strings, names a schema that another context owns.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d76da72765f05178b2aa0d251afd406e8b9c66929df82453136cd27cbac83d7a
semantic_inventory:
  - claim_key: CLAIM-14088189B2AC0C8B
    claim_text: The architecture tests must fail when a context's SQL, in its migrations or in its packages' query strings, names a schema that another context owns
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:49.743Z'
id: REQ-tooling-archtest-schemas
type: req
---
The architecture tests must fail when a context's SQL, in its migrations or in its packages' query strings, names a schema that another context owns.

## Context

Each context owns its schema (code-organization.md, P). The River job queue's schema river belongs to valuation, which migrates and runs it (D-36). SQL comments do not count. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> A context's SQL, in its migrations and its adapters' queries, names only its own schema.

docs/dev/code-organization.md#architecture-tests
