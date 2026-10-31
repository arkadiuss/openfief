---
title: The docs site published from main must be built only after both end-to-end runs passed, with each run's kibi.proof-run.v1 artifact ingested by kibi into the same commit's knowledge base, oldest run first, and must fail its build when kibi refuses an artifact.
status: open
text_ref: docs/dev/kibi.md#evidence-as-kibi-counts-it
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-evidence-as-kibi-counts-it
  - lane:strict
semantic_text: The docs site published from main must be built only after both end-to-end runs passed, with each run's kibi.proof-run.v1 artifact ingested by kibi into the same commit's knowledge base, oldest run first, and must fail its build when kibi refuses an artifact.
logic_claims:
  - CLAIM-41C13B43A8C5CDA1
semantic_clauses:
  - The docs site published from main must be built only after both end-to-end runs passed, with each run's kibi.proof-run.v1 artifact ingested by kibi into the same commit's knowledge base, oldest run first, and must fail its build when kibi refuses an artifact.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f1af72914183c9aad55ed85dc07aab60f2d35681be1c2c357d15f9b50a95c813
semantic_inventory:
  - claim_key: CLAIM-41C13B43A8C5CDA1
    claim_text: The docs site published from main must be built only after both end-to-end runs passed, with each run's kibi.proof-run.v1 artifact ingested by kibi into the same commit's knowledge base, oldest run first, and must fail its build when kibi refuses an artifact
    role: normative
    status: modeled
    span:
      start: 0
      end: 258
origin:
  kind: agent
  recorded_at: '2026-10-31T12:45:48.881Z'
id: REQ-tooling-docs-report-e2e-receipts
type: req
---
The docs site published from main must be built only after both end-to-end runs passed, with each run's kibi.proof-run.v1 artifact ingested by kibi into the same commit's knowledge base, oldest run first, and must fail its build when kibi refuses an artifact.

## Context

The author decided that kibi counts a requirement as tested only with a fresh, passing end-to-end receipt from kibi prove, that CI's end-to-end job runs kibi prove, and that the requirements report the site publishes from main shows that run's receipts. The e2e jobs run on k3d and kind; kibi keeps a test's receipts in the order their runs finished, so they are ingested oldest first. Requested by the author through the orchestrating agent; modeled by Claude Code, not yet reviewed.

## Source

> CI's end-to-end job runs `kibi prove`, and the requirements report the site publishes from `main` shows that run's receipts.

docs/dev/kibi.md#evidence-as-kibi-counts-it
