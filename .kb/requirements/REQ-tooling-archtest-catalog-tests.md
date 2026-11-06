---
title: 'The architecture tests must fail when a test pins the embedded plugin catalog: it compares the number of catalog.Load() entries with a literal, picks an entry by a literal index, spells out a digest the catalog pins, or inspects what a catalog-reading function of its package returns without deriving its expectation from catalog.Load().'
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: 'The architecture tests must fail when a test pins the embedded plugin catalog: it compares the number of catalog.Load() entries with a literal, picks an entry by a literal index, spells out a digest the catalog pins, or inspects what a catalog-reading function of its package returns without deriving its expectation from catalog.Load().'
logic_claims:
  - CLAIM-403971453AC7E999
semantic_clauses:
  - 'The architecture tests must fail when a test pins the embedded plugin catalog: it compares the number of catalog.Load() entries with a literal, picks an entry by a literal index, spells out a digest the catalog pins, or inspects what a catalog-reading function of its package returns without deriving its expectation from catalog.Load().'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 60a1c805ddfa4701a247eeb041e4f52ae07d1f252198bffdb9265741fac80a1f
semantic_inventory:
  - claim_key: CLAIM-403971453AC7E999
    claim_text: 'The architecture tests must fail when a test pins the embedded plugin catalog: it compares the number of catalog.Load() entries with a literal, picks an entry by a literal index, spells out a digest the catalog pins, or inspects what a catalog-reading function of its package returns without deriving its expectation from catalog.Load()'
    role: normative
    status: modeled
    span:
      start: 0
      end: 336
origin:
  kind: agent
  recorded_at: '2026-11-06T17:50:02.873Z'
id: REQ-tooling-archtest-catalog-tests
type: req
---
The architecture tests must fail when a test pins the embedded plugin catalog: it compares the number of catalog.Load() entries with a literal, picks an entry by a literal index, spells out a digest the catalog pins, or inspects what a catalog-reading function of its package returns without deriving its expectation from catalog.Load().

## Context

The embedded catalog grows with every release (D-88). After the 2026.11.0-alpha.1 release added seven entries, a test that expected a policy built from an empty catalog broke main; the design session asked for an architecture rule against tests that pin the catalog. Checking only a catalog reader's error is allowed. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> The embedded plugin catalog grows with every release, so no test pins it.

docs/dev/code-organization.md#architecture-tests
