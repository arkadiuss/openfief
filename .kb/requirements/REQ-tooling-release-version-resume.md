---
title: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists, except that the publish step accepts a tag that already points at the release commit it publishes.
status: open
text_ref: docs/dev/pipeline.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-releases
  - lane:strict
semantic_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists, except that the publish step accepts a tag that already points at the release commit it publishes.
logic_claims:
  - CLAIM-142ADD621CAB56E3
semantic_clauses:
  - The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists, except that the publish step accepts a tag that already points at the release commit it publishes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 49478303fde04859ae8985757f0299f50d00f44946b963636806dcea6945ef1e
semantic_inventory:
  - claim_key: CLAIM-142ADD621CAB56E3
    claim_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists, except that the publish step accepts a tag that already points at the release commit it publishes
    role: exception
    status: modeled
    span:
      start: 0
      end: 338
origin:
  kind: agent
  recorded_at: '2026-11-05T21:54:30.274Z'
id: REQ-tooling-release-version-resume
type: req
---
The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists, except that the publish step accepts a tag that already points at the release commit it publishes.

## Context

The first release, 2026.11.0-alpha.1 (run 42105814769), created its tag and published the pre-release, then failed at gh pr create because the repository did not yet allow GitHub Actions to open pull requests. Re-running the failed jobs then stopped at release:check with 'tag already exists', so the run could never finish and deploy demo was skipped. Requested in the release-resumability brief (fix/release-resumable). The plan job still refuses any existing tag: a fresh run would make a different catalog commit.

## Why this changed

Evidence: run 42105814769 of the 2026.11.0-alpha.1 release created the tag at the release commit and failed afterwards, and re-running the failed jobs could not resume because the superseded rule refused every existing tag, including the one this very release created. A tag at any other commit is still refused.

## Source

> Given a release commit (`COMMIT`, which only `release:publish` passes), a tag already at that commit passes too, so a publish can resume; a tag anywhere else is still refused.

docs/dev/pipeline.md#releases
