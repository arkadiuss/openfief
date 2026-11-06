---
title: The architecture tests must fail when a logging call takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type, or an error that is not passed through a LogSafe function.
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: The architecture tests must fail when a logging call takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type, or an error that is not passed through a LogSafe function.
logic_claims:
  - CLAIM-F41C06A280383C4C
semantic_clauses:
  - The architecture tests must fail when a logging call takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type, or an error that is not passed through a LogSafe function.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9dad2732aa54194f4949bafb7226ee02ba8adab42007f88e7de4f657f1cb2fe3
semantic_inventory:
  - claim_key: CLAIM-F41C06A280383C4C
    claim_text: The architecture tests must fail when a logging call takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type, or an error that is not passed through a LogSafe function
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:56.471Z'
id: REQ-tooling-archtest-privacy
type: req
---
The architecture tests must fail when a logging call takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type, or an error that is not passed through a LogSafe function.

## Context

GR-12 and code-organization.md: never log amounts, holdings, account numbers or tokens. An error's text may quote such a value, so errors are logged through kernel.LogSafe or a context's LogSafe. Strings holding account numbers have no type to tell them apart and stay with review. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> No logging call takes a value of a money, quantity, price or rate type, an ISIN, a broker account number or a token, as an argument or inside an error it logs.

docs/dev/code-organization.md#architecture-tests
