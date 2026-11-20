---
title: Tooling Release Smoke
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: tooling.release_smoke
canonical_key: tooling.release_smoke
claim_key: CLAIM-F62483E029A26DC6
claim_text: After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says.
origin:
  kind: agent
  recorded_at: '2026-11-06T16:42:34.986Z'
id: FACT-SUBJECT-FD8AACF94378B2E4
type: fact
---

Definition not stated: the subject tooling.release_smoke is named by requirement REQ-tooling-release-smoke, which links to the subject fact FACT-SUBJECT-FD8AACF94378B2E4 through constrains. The people who own that source can say what it is in the product and what it is not.
