---
title: Ui Offline Overview
status: active
text_ref: design-decisions:D-151#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-151-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ui.offline_overview
canonical_key: ui.offline_overview
claim_key: CLAIM-7ABDF5913E9D3770
claim_text: 'The installed app must keep only the last overview on the device for offline reading: totals, each account''s values and the allocation''s shares, with the time they''re from.'
origin:
  kind: agent
  recorded_at: '2026-11-01T12:28:21.642Z'
id: FACT-SUBJECT-6AC55871A4F16CF6
type: fact
---

Definition not stated: the subject ui.offline_overview is named by requirements REQ-ui-offline-keeps-last-overview-only, REQ-ui-offline-never-keeps-records, REQ-ui-offline-overview-contents, REQ-ui-offline-overview-unexportable-key, REQ-ui-offline-overview-wipe-triggers, REQ-ui-offline-overview-wiped-on-session-end, which link to the subject fact FACT-SUBJECT-6AC55871A4F16CF6 through constrains. The people who own that source can say what it is in the product and what it is not.
