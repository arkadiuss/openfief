---
title: Ingest Ticker Entries
status: active
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.ticker_entries
canonical_key: ingest.ticker_entries
claim_key: CLAIM-CA0492CD5AB6399A
claim_text: A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:17.971Z'
id: FACT-SUBJECT-8A8B03F40D5773D9
type: fact
---

Definition not stated: the subject ingest.ticker_entries is named by requirements REQ-ingest-ticker-confirmation, REQ-ingest-ticker-contract, REQ-ingest-ticker-currency-picked, REQ-ingest-ticker-entries-wait, REQ-ingest-ticker-isin-lookup, REQ-ingest-ticker-opening-lots, REQ-ingest-ticker-trading-currency, REQ-ingest-xtb-opening-lot-cost-trading-currency, which link to the subject fact FACT-SUBJECT-8A8B03F40D5773D9 through constrains. The people who own that source can say what it is in the product and what it is not.
