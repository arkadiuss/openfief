---
title: Ingest Multi Account Upload
status: active
text_ref: design-decisions:D-140#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-140-1
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: ingest.multi_account_upload
canonical_key: ingest.multi_account_upload
claim_key: CLAIM-FF05AFDD4B3E5FE8
claim_text: An upload of a statement covering several accounts must run in the mode chosen for the account it was uploaded for, which the owner can override.
origin:
  kind: agent
  recorded_at: '2026-11-01T12:23:40.286Z'
id: FACT-SUBJECT-B4872D883DCCB3F1
type: fact
---

Definition not stated: the subject ingest.multi_account_upload is named by requirements REQ-ingest-multi-account-opening-seeds-unseeded-only, REQ-ingest-multi-account-transactions-books-every-mapped, REQ-ingest-multi-account-upload-changes-every-bound, REQ-ingest-multi-account-upload-mode-of-uploaded-account, REQ-ingest-statement-multi-account-reports-unmapped, which link to the subject fact FACT-SUBJECT-B4872D883DCCB3F1 through constrains. The people who own that source can say what it is in the product and what it is not.
