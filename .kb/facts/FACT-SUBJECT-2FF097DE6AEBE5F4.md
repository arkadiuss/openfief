---
title: Ingest Xtb Holding Assertions
status: active
text_ref: design-decisions:D-155#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-155-1
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: ingest.xtb_holding_assertions
canonical_key: ingest.xtb_holding_assertions
claim_key: CLAIM-E87C693F776FB8EA
claim_text: 'An XTB export''s Open Positions sheet must also assert the account''s holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report''s date.'
origin:
  kind: agent
  recorded_at: '2026-11-01T12:30:01.512Z'
id: FACT-SUBJECT-2FF097DE6AEBE5F4
type: fact
---

Definition not stated: the subject ingest.xtb_holding_assertions is named by requirements REQ-ingest-xtb-open-positions-assert-holdings, REQ-ingest-xtb-unconfirmed-ticker-asserted-on-confirm, which link to the subject fact FACT-SUBJECT-2FF097DE6AEBE5F4 through constrains. The people who own that source can say what it is in the product and what it is not.
