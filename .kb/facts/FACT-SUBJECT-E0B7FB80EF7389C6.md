---
title: Release Publish
status: active
text_ref: docs/dev/pipeline.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-releases
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: release.publish
canonical_key: release.publish
claim_key: CLAIM-7256562FE856A499
claim_text: 'Re-running a release publish that failed part-way must finish only what is missing: keep a tag already at the release commit, create the release only when it does not exist and otherwise upload only the assets it lacks by name, and open the catalog pull request only when no open or merged one from release/<version> exists.'
origin:
  kind: agent
  recorded_at: '2026-11-05T21:54:50.072Z'
id: FACT-SUBJECT-E0B7FB80EF7389C6
type: fact
---
