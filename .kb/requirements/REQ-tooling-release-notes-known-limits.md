---
title: When the release ref has .github/release-notes/<version>/known-limits.md, the release notes must include its Markdown as a Known limits section ahead of the changes, and an empty file must fail the notes.
status: open
text_ref: docs/dev/pipeline.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-releases
  - lane:strict
semantic_text: When the release ref has .github/release-notes/<version>/known-limits.md, the release notes must include its Markdown as a Known limits section ahead of the changes, and an empty file must fail the notes.
logic_claims:
  - CLAIM-04AD771AB90B81FF
semantic_clauses:
  - When the release ref has .github/release-notes/<version>/known-limits.md, the release notes must include its Markdown as a Known limits section ahead of the changes, and an empty file must fail the notes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c70242608303979b0f43ccbaa4ebda2a0c2d8495002d0cee2dd10b81be107fb1
semantic_inventory:
  - claim_key: CLAIM-04AD771AB90B81FF
    claim_text: When the release ref has .github/release-notes/<version>/known-limits.md, the release notes must include its Markdown as a Known limits section ahead of the changes, and an empty file must fail the notes
    role: condition
    status: modeled
    span:
      start: 0
      end: 203
origin:
  kind: agent
  recorded_at: '2026-11-05T21:55:38.456Z'
id: REQ-tooling-release-notes-known-limits
type: req
---
When the release ref has .github/release-notes/<version>/known-limits.md, the release notes must include its Markdown as a Known limits section ahead of the changes, and an empty file must fail the notes.

## Context

The 2026.11.0-alpha.1 notes needed a Known limits section that the generated notes could not hold. Future releases take it from a per-version file the author commits to main before starting the release, read at the release ref so the dry run and the publish agree. Requested in the release-resumability brief (fix/release-resumable); the path is Claude Code's choice.

## Source

> A release's known limits go in `.github/release-notes/<version>/known-limits.md`, committed to `main` before the release starts: when the release commit has that file, its Markdown, the section's body without a heading, becomes a Known limits section ahead of the changes.

docs/dev/pipeline.md#releases
