---
title: Loads Through Real Paths = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.loader
property_key: loads_through_real_paths
canonical_key: mcp-kibi-model-requirement:demo.loader:loads_through_real_paths:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-39974F0AD4840B58
claim_text: 'The demo loader must load the dataset through OpenFief''s real paths: the IBKR history as Flex statements through the ibkrflex connector and the ingest service, seeded at go-live and then one transactions batch per month; the XTB and bond accounts through the ledger''s seeding and import use cases; prices, FX rates, CPI, bond terms and the IKE wrapper through valuation''s input recorder; and the stored valuations through the River revaluation workers, exiting once every demo account is valued through today.'
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:28.690Z'
id: FACT-PROP-FE780A5DEAE4E1AB
type: fact
---
