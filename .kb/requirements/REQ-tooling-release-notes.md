---
title: Release notes must list the commits since the previous release grouped by Conventional Commit type, and a stable release must start after the previous stable release.
status: open
text_ref: docs/dev/sdlc.md#commits-and-prs
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-commits-and-prs
  - lane:strict
semantic_text: Release notes must list the commits since the previous release grouped by Conventional Commit type, and a stable release must start after the previous stable release.
logic_claims:
  - CLAIM-993F1B2D4979E168
semantic_clauses:
  - Release notes must list the commits since the previous release grouped by Conventional Commit type, and a stable release must start after the previous stable release.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: db96d2b445dd7d4939a34bbe85ac1c5b73591b818757d9c5469045a2f4b9cb4e
semantic_inventory:
  - claim_key: CLAIM-993F1B2D4979E168
    claim_text: Release notes must list the commits since the previous release grouped by Conventional Commit type, and a stable release must start after the previous stable release
    role: normative
    status: modeled
    span:
      start: 0
      end: 165
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:57.806Z'
id: REQ-tooling-release-notes
type: req
---
Release notes must list the commits since the previous release grouped by Conventional Commit type, and a stable release must start after the previous stable release.

## Context

docs/dev/sdlc.md proposes (P) generating release notes from Conventional Commits. A small script was chosen over GitHub's generated notes, which group by pull request labels rather than commit types, and over an extra pinned tool, so the dry run and a local run produce the same notes with no new dependency. Breaking changes come first; a stable release covers everything since the previous stable, including its betas. Requested in the SDLC/CI implementation brief.

## Source

> Conventional Commits; release notes are generated from them **(P)**.

docs/dev/sdlc.md#commits-and-prs
