---
title: The architecture tests must fail when a test calls time.Sleep or skips itself, unless the exceptions list that test with a reason.
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: The architecture tests must fail when a test calls time.Sleep or skips itself, unless the exceptions list that test with a reason.
logic_claims:
  - CLAIM-21ACD3A725C8AF19
semantic_clauses:
  - The architecture tests must fail when a test calls time.Sleep or skips itself, unless the exceptions list that test with a reason.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a00252bf1e099c7ac00c4e9e60647b935a2e02bfec5fa126a3f7bf7b2c86a10c
semantic_inventory:
  - claim_key: CLAIM-21ACD3A725C8AF19
    claim_text: The architecture tests must fail when a test calls time.Sleep or skips itself, unless the exceptions list that test with a reason
    role: exception
    status: modeled
    span:
      start: 0
      end: 129
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:59.671Z'
id: REQ-tooling-archtest-stable-tests
type: req
---
The architecture tests must fail when a test calls time.Sleep or skips itself, unless the exceptions list that test with a reason.

## Context

testing.md: no sleeps, wait on conditions with a deadline, and never skip to get green. A few tests wait on something outside the process or need a tool a contributor may lack, such as PostgreSQL, and are listed with their reasons. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> No time.Sleep and no t.Skip in tests, outside a listed few that wait on something outside the process or need a tool a contributor may lack, each with its reason.

docs/dev/code-organization.md#architecture-tests
