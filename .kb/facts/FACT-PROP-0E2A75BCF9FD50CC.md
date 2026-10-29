---
title: Reported By Connector And Picked In Ui = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-72-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-72#2
fact_kind: property_value
subject_key: ingest.broker_account
property_key: reported_by_connector_and_picked_in_ui
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:ingest.broker_account:reported_by_connector_and_picked_in_ui:eq:true
claim_key: CLAIM-73CB3E1460FB804A
claim_text: A broker connector must report the broker accounts it saw on its last sync, and the UI must offer them to pick from, so that the owner never types a broker account number.
origin:
  kind: agent
  recorded_at: '2026-10-29T05:38:05.553Z'
id: FACT-PROP-0E2A75BCF9FD50CC
type: fact
---
