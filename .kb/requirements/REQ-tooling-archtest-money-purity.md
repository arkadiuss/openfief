---
title: The architecture tests must fail when a domain, app or kernel package uses float32 or float64, when a contract under proto/ has a float or double field, or when a domain or kernel package starts a goroutine or holds a package-level variable other than an error.
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: The architecture tests must fail when a domain, app or kernel package uses float32 or float64, when a contract under proto/ has a float or double field, or when a domain or kernel package starts a goroutine or holds a package-level variable other than an error.
logic_claims:
  - CLAIM-0EF278FC493C9404
semantic_clauses:
  - The architecture tests must fail when a domain, app or kernel package uses float32 or float64, when a contract under proto/ has a float or double field, or when a domain or kernel package starts a goroutine or holds a package-level variable other than an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c4f1ac719ab770394cb0826137cfa64d157c5bbea434badb75dc89c6f6324615
semantic_inventory:
  - claim_key: CLAIM-0EF278FC493C9404
    claim_text: The architecture tests must fail when a domain, app or kernel package uses float32 or float64, when a contract under proto/ has a float or double field, or when a domain or kernel package starts a goroutine or holds a package-level variable other than an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 260
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:53.123Z'
id: REQ-tooling-archtest-money-purity
type: req
---
The architecture tests must fail when a domain, app or kernel package uses float32 or float64, when a contract under proto/ has a float or double field, or when a domain or kernel package starts a goroutine or holds a package-level variable other than an error.

## Context

AGENTS.md and code-organization.md forbid float64 for money, and keep the domain pure; lint bans imports such as sync and os in the domain but cannot see a go statement or package state. Packages that only draw may be listed as exceptions. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> No float32 or float64 in domain, app and kernel packages, and no float or double field in the contracts under proto/.

docs/dev/code-organization.md#architecture-tests
