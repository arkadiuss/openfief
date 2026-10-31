---
title: Installing a connector whose manifest declares it a market-data connector must show the owner that the plugin will learn what they hold.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-123-2
  - lane:strict
text_ref: design-decisions:D-123#2
origin:
  kind: import
  ref: design-decisions:D-123#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Installing a connector whose manifest declares it a market-data connector must show the owner that the plugin will learn what they hold.
logic_claims:
  - CLAIM-CC3C50FBDE2FE6B3
semantic_clauses:
  - Installing a connector whose manifest declares it a market-data connector must show the owner that the plugin will learn what they hold.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4254f58525cab9947eefd46de8f80aeb96713ebfb4f40b6e970f564c85ce0d20
semantic_inventory:
  - claim_key: CLAIM-CC3C50FBDE2FE6B3
    claim_text: Installing a connector whose manifest declares it a market-data connector must show the owner that the plugin will learn what they hold
    role: normative
    status: modeled
    span:
      start: 0
      end: 135
id: REQ-plugins-install-shows-market-data-list
type: req
---
Installing a connector whose manifest declares it a market-data connector must show the owner that the plugin will learn what they hold.

## Context

D-93's list of held instruments and model targets tells a market-data connector what the owner holds at every institution, and D-123 sends it only to a connector whose manifest declares it a market-data connector and whose grant approves that (REQ-ingest-market-data-list-declared-granted). The author gave the reason in the same decision: because the declaration is approved with the grant, the install shows the owner, before they approve it, that the plugin will learn what they hold. The implementation raised the question of which connectors get the list as question #7 on docs/implementation-questions.md, and the author decided it on 2026-10-31.

## Source

> Only a connector whose manifest declares it a market-data connector, and whose grant approves that, receives D-93's list of held instruments and model targets, so the install shows the owner that the plugin will learn what they hold.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-123#2
