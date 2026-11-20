---
title: Tooling Breaking Change Exceptions
status: active
text_ref: docs/dev/pipeline.md#breaking-contract-changes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-breaking-contract-changes
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: tooling.breaking_change_exceptions
canonical_key: tooling.breaking_change_exceptions
claim_key: CLAIM-0B50F30D5BCAF230
claim_text: A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it.
origin:
  kind: agent
  recorded_at: '2026-11-01T19:21:39.392Z'
id: FACT-SUBJECT-C87F01A26ABBB0C2
type: fact
---

Definition not stated: the subject tooling.breaking_change_exceptions is named by requirements REQ-tooling-breaking-before-first-beta, REQ-tooling-breaking-exception-commit-marked, which link to the subject fact FACT-SUBJECT-C87F01A26ABBB0C2 through constrains. The people who own that source can say what it is in the product and what it is not.
