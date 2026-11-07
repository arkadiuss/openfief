---
title: Outside CI, the docs build must reuse a kibi coverage report it kept for the same working-tree contents of the .kb files, installed kibi-cli and kibi-core versions and local kibi patches, and say so; otherwise, and always in CI, it must ask kibi for a fresh report.
status: open
text_ref: docs/dev/docs.md#build-and-publishing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md-build-and-publishing
  - lane:strict
semantic_text: Outside CI, the docs build must reuse a kibi coverage report it kept for the same working-tree contents of the .kb files, installed kibi-cli and kibi-core versions and local kibi patches, and say so; otherwise, and always in CI, it must ask kibi for a fresh report.
logic_claims:
  - CLAIM-4637F5D6C69CBEA1
semantic_clauses:
  - Outside CI, the docs build must reuse a kibi coverage report it kept for the same working-tree contents of the .kb files, installed kibi-cli and kibi-core versions and local kibi patches, and say so; otherwise, and always in CI, it must ask kibi for a fresh report.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2bbe153851cf4176a9fb510c91e8735f34eb464e5b9470d000d956c15fdf66ee
semantic_inventory:
  - claim_key: CLAIM-4637F5D6C69CBEA1
    claim_text: Outside CI, the docs build must reuse a kibi coverage report it kept for the same working-tree contents of the .kb files, installed kibi-cli and kibi-core versions and local kibi patches, and say so; otherwise, and always in CI, it must ask kibi for a fresh report
    role: normative
    status: modeled
    span:
      start: 0
      end: 264
origin:
  kind: agent
  recorded_at: '2026-11-07T21:14:26.428Z'
id: REQ-tooling-kibi-coverage-cache
type: req
---
Outside CI, the docs build must reuse a kibi coverage report it kept for the same working-tree contents of the .kb files, installed kibi-cli and kibi-core versions and local kibi patches, and say so; otherwise, and always in CI, it must ask kibi for a fresh report.

## Context

Paged, the coverage report takes about four minutes, which the pre-commit hook paid on every commit touching docs. The report is generated site content fully determined by its inputs, so reusing it is not a weaker check: kibi check and kibi sync still run as before, and CI, which builds the published site, always regenerates it. Requested by the author through the design session; modeled by Claude Code, not yet reviewed.

## Source

> Cache the generated coverage so the commit hook doesn't pay 3-4 min on every docs commit: key = hash of the tracked .kb/ tree contents + the installed kibi-cli and kibi-core versions + both patch files ... CI's docs-build job must ALWAYS regenerate from scratch (no cache).

docs/dev/docs.md#build-and-publishing
