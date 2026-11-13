---
title: The pre-commit kibi step must keep the throwaway knowledge base that kibi check --staged builds off the disk while it projects the branch's records into it, and must remove the throwaway knowledge bases that killed runs left in the temporary directory before it runs the checks, so that a one-file commit's staged check does not grow with the square of the knowledge base.
status: open
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
semantic_text: The pre-commit kibi step must keep the throwaway knowledge base that kibi check --staged builds off the disk while it projects the branch's records into it, and must remove the throwaway knowledge bases that killed runs left in the temporary directory before it runs the checks, so that a one-file commit's staged check does not grow with the square of the knowledge base.
logic_claims:
  - CLAIM-30DFEB40D9896388
semantic_clauses:
  - The pre-commit kibi step must keep the throwaway knowledge base that kibi check --staged builds off the disk while it projects the branch's records into it, and must remove the throwaway knowledge bases that killed runs left in the temporary directory before it runs the checks, so that a one-file commit's staged check does not grow with the square of the knowledge base.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4c99c8ae566c4eba051862fdf01dab7b3a0f3a36e6229a2e6f3bee5a2e2694f4
semantic_inventory:
  - claim_key: CLAIM-30DFEB40D9896388
    claim_text: The pre-commit kibi step must keep the throwaway knowledge base that kibi check --staged builds off the disk while it projects the branch's records into it, and must remove the throwaway knowledge bases that killed runs left in the temporary directory before it runs the checks, so that a one-file commit's staged check does not grow with the square of the knowledge base
    role: normative
    status: modeled
    span:
      start: 0
      end: 371
origin:
  kind: agent
  recorded_at: '2026-11-13T02:54:35.154Z'
id: REQ-tooling-kibi-staged-store-unsaved
type: req
---
The pre-commit kibi step must keep the throwaway knowledge base that kibi check --staged builds off the disk while it projects the branch's records into it, and must remove the throwaway knowledge bases that killed runs left in the temporary directory before it runs the checks, so that a one-file commit's staged check does not grow with the square of the knowledge base.

## Context

`kibi check --staged` took 15 minutes and more for a one-file commit once the knowledge base reached about 14,000 symbols, because kibi saved its whole throwaway store after each batch of 128 records it projected, and killed runs left 31 such stores (0.5 GB) in the temporary directory. Found and requested by the author through the pipeline-stability work; the fix is the kibi-cli patch and scripts/kibi-staged.sh. Modeled by Claude Code, not yet reviewed.

## Source

> Pipeline stability is top priority: kibi:staged takes ~1000-1300 s even when only one test file is staged ... fix it properly, without weakening the gate. Target: kibi:staged under ~60 s for a one-file commit.
