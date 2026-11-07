---
title: Egress Only Configured Destinations = true
status: active
text_ref: design-decisions:D-200#16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-16
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: egress_only_configured_destinations
canonical_key: mcp-kibi-model-requirement:m0.exit:egress_only_configured_destinations:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-52E8C5F1BF11D050
claim_text: Outbound traffic must go only to configured destinations, such as IBKR's Flex service, NBP, GUS, the market-data vendor and the user's Ghostfolio, as verified from the egress proxy's logs.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:32:37.072Z'
id: FACT-PROP-5D7AF718DEEEF5EF
type: fact
---
