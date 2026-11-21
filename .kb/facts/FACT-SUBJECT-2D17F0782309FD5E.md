---
title: Identity Read Access
status: active
text_ref: design-decisions:D-137#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.read_access
canonical_key: identity.read_access
claim_key: CLAIM-A64525E68682D0DE
claim_text: 'Read access must be everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins.'
origin:
  kind: agent
  recorded_at: '2026-11-01T07:40:01.368Z'
id: FACT-SUBJECT-2D17F0782309FD5E
type: fact
---

Definition not stated: the subject identity.read_access is named by requirements REQ-identity-read-access-holders, REQ-identity-read-access-holders-incl-owner, REQ-identity-read-access-kinds-of-record, REQ-identity-read-access-never-broker-account-numbers, REQ-identity-read-access-never-secrets, REQ-identity-read-access-never-state-change, REQ-identity-read-access-no-broker-account-numbers, REQ-identity-read-access-no-secrets, REQ-identity-read-access-no-state-change, REQ-identity-read-access-scope, REQ-identity-read-access-viewable-accounts-only, which link to the subject fact FACT-SUBJECT-2D17F0782309FD5E through constrains. The people who own that source can say what it is in the product and what it is not.
