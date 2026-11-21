---
title: Plugins Manifest
status: active
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.manifest
canonical_key: plugins.manifest
claim_key: CLAIM-75F75192E8FF565B
claim_text: A manifest must declare at least one known capability and valid, unique egress hosts, Secrets, mount paths and data tiers, and validation must report every invalid field by name.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:34.597Z'
id: FACT-SUBJECT-777F1580F1EE5686
type: fact
---
A plugin's signed manifest: embedded in its image as a label on the image config, and declaring its capabilities, egress hosts, Secrets and their roles, mount paths, statement files and data tiers, validated so that every invalid field is reported by name. It does not cover the grant that approves what the manifest requests.