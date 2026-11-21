---
title: Views Read Api
status: active
text_ref: docs/DESIGN.md#D-50
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-50
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: views.read_api
canonical_key: views.read_api
claim_key: CLAIM-F8D92AA3C4096CAA
claim_text: 'The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal.'
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:34.353Z'
id: FACT-SUBJECT-5D5273B6E7DF2D96
type: fact
---

Definition not stated: the subject views.read_api is named by requirements REQ-views-api-auth, REQ-views-api-contract, REQ-views-api-logs, REQ-views-api-read-only, REQ-views-authz-hidden-account-not-found, REQ-views-authz-totals-over-viewable, REQ-views-bond-series-terms, REQ-views-model-off-404, REQ-views-model-portfolio, REQ-views-onboarding-lists, REQ-views-price-checks, REQ-views-read-api-bounded-bodies, REQ-views-read-api-declared-errors, REQ-views-read-api-list-broker-accounts-removed, REQ-views-read-api-no-broker-account-numbers, REQ-views-read-api-serves-no-broker-accounts, which link to the subject fact FACT-SUBJECT-5D5273B6E7DF2D96 through constrains. The people who own that source can say what it is in the product and what it is not.
