---
title: Ingest Connector Contract Messages
status: active
text_ref: docs/dev/code-organization.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.connector_contract_messages
canonical_key: ingest.connector_contract_messages
claim_key: CLAIM-A7EAE16BB38054D2
claim_text: The core must translate connector contract messages into domain values in its Connect adapter, refusing, with the field named, a decimal with an exponent or grouping, a date that is not YYYY-MM-DD, a malformed currency or ISIN, a missing required field, an unspecified enum or a delivery for another batch; no contract type may reach the domain or the use cases.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:48.174Z'
id: FACT-SUBJECT-69A02BF8F0306433
type: fact
---

Definition not stated: the subject ingest.connector_contract_messages is named by requirement REQ-ingest-connector-contract-translation, which links to the subject fact FACT-SUBJECT-69A02BF8F0306433 through constrains. The people who own that source can say what it is in the product and what it is not.
