---
title: 'Re-running a release publish that failed part-way must finish only what is missing: keep a tag already at the release commit, create the release only when it does not exist and otherwise upload only the assets it lacks by name, and open the catalog pull request only when no open or merged one from release/<version> exists.'
status: open
text_ref: docs/dev/pipeline.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-releases
  - lane:strict
semantic_text: 'Re-running a release publish that failed part-way must finish only what is missing: keep a tag already at the release commit, create the release only when it does not exist and otherwise upload only the assets it lacks by name, and open the catalog pull request only when no open or merged one from release/<version> exists.'
logic_claims:
  - CLAIM-7256562FE856A499
semantic_clauses:
  - 'Re-running a release publish that failed part-way must finish only what is missing: keep a tag already at the release commit, create the release only when it does not exist and otherwise upload only the assets it lacks by name, and open the catalog pull request only when no open or merged one from release/<version> exists.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b9847dfcc3fbbeda800a2df39ca43cb630b26b30a8c7ba0b477b58239e6238af
semantic_inventory:
  - claim_key: CLAIM-7256562FE856A499
    claim_text: 'Re-running a release publish that failed part-way must finish only what is missing: keep a tag already at the release commit, create the release only when it does not exist and otherwise upload only the assets it lacks by name, and open the catalog pull request only when no open or merged one from release/<version> exists'
    role: normative
    status: modeled
    span:
      start: 0
      end: 323
origin:
  kind: agent
  recorded_at: '2026-11-05T21:54:52.282Z'
id: REQ-tooling-release-publish-resumable
type: req
---
Re-running a release publish that failed part-way must finish only what is missing: keep a tag already at the release commit, create the release only when it does not exist and otherwise upload only the assets it lacks by name, and open the catalog pull request only when no open or merged one from release/<version> exists.

## Context

The 2026.11.0-alpha.1 release run 42105814769 published the tag and the pre-release with all 17 assets, then failed at gh pr create because the repository did not allow GitHub Actions to create pull requests; re-running the failed jobs could not finish and deploy demo never ran. Requested in the release-resumability brief (fix/release-resumable). An asset whose upload never finished counts as missing.

## Source

> Publishing is resumable. When `publish` fails part-way, **Re-run failed jobs** runs `publish` and then `deploy-demo` again on the same release commit, since the earlier jobs' outputs are kept.

docs/dev/pipeline.md#releases
