---
title: The architecture tests must fail on any violation that neither the baseline nor the exceptions list, and on any listed line that no longer matches a violation, so the baseline only shrinks and both lists stay exact.
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: The architecture tests must fail on any violation that neither the baseline nor the exceptions list, and on any listed line that no longer matches a violation, so the baseline only shrinks and both lists stay exact.
logic_claims:
  - CLAIM-E35CA450D9E0E5E1
semantic_clauses:
  - The architecture tests must fail on any violation that neither the baseline nor the exceptions list, and on any listed line that no longer matches a violation, so the baseline only shrinks and both lists stay exact.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9e32acf804a54953f738a6844a2200ab6b8445cf7aa02785655561f34a0f8b2e
semantic_inventory:
  - claim_key: CLAIM-E35CA450D9E0E5E1
    claim_text: The architecture tests must fail on any violation that neither the baseline nor the exceptions list, and on any listed line that no longer matches a violation, so the baseline only shrinks and both lists stay exact
    role: normative
    status: modeled
    span:
      start: 0
      end: 214
origin:
  kind: agent
  recorded_at: '2026-11-06T17:50:06.262Z'
id: REQ-tooling-archtest-baseline
type: req
---
The architecture tests must fail on any violation that neither the baseline nor the exceptions list, and on any listed line that no longer matches a violation, so the baseline only shrinks and both lists stay exact.

## Context

code-organization.md: a rule the code breaks today starts with its breaks in a checked-in baseline, each with its reason, and the list only shrinks; it is never edited to get green. Each line names a stable key, never a line number. The design session decided the ingest to portfolio import starts in the baseline until an ingest/app port wired in cmd/core replaces it. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> The test fails on any break the baseline doesn't list and on any listed break that no longer happens, so the list only shrinks and is never edited to get green.

docs/dev/code-organization.md#architecture-tests
