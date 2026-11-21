---
title: Plugins Plugin Upgrade
status: archived
text_ref: design-decisions:D-16#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-16-3
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: plugins.plugin_upgrade
canonical_key: plugins.plugin_upgrade
claim_key: CLAIM-4E50A2E1D2C2CAD3
claim_text: A plugin upgrade that requests more permissions must not deploy until they are approved.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:30.416Z'
id: FACT-SUBJECT-60D89A948A654996
type: fact
part_of:
  - plugins.upgrade
---
The approval gate of a plugin upgrade: one that requests more permissions must not deploy until they are approved. It is a part of plugins.upgrade, which also covers the report of the permissions a grant does not approve.