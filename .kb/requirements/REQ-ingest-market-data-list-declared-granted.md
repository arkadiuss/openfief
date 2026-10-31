---
title: Only a connector whose manifest declares it a market-data connector, and whose grant approves that, must receive D-93's list of held instruments and model targets.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-123-1
  - lane:strict
text_ref: design-decisions:D-123#1
origin:
  kind: import
  ref: design-decisions:D-123#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Only a connector whose manifest declares it a market-data connector, and whose grant approves that, must receive D-93's list of held instruments and model targets.
logic_claims:
  - CLAIM-19E87FB1C8E875E6
semantic_clauses:
  - Only a connector whose manifest declares it a market-data connector, and whose grant approves that, must receive D-93's list of held instruments and model targets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7bda23181349bce041748fc5ecdf57c6183b0d47a8cf2f1eee14e6a060105557
semantic_inventory:
  - claim_key: CLAIM-19E87FB1C8E875E6
    claim_text: Only a connector whose manifest declares it a market-data connector, and whose grant approves that, must receive D-93's list of held instruments and model targets
    role: normative
    status: modeled
    span:
      start: 0
      end: 162
id: REQ-ingest-market-data-list-declared-granted
type: req
---
Only a connector whose manifest declares it a market-data connector, and whose grant approves that, must receive D-93's list of held instruments and model targets.

## Context

D-93 has each sync request to a market-data connector carry the instruments the owner holds and the model portfolio's targets (REQ-ingest-market-data-sync-instruments), but no manifest field said which connector is a market-data connector, and the list tells a plugin what the owner holds at every institution. The implementation sends it to every connector instance bound to no account (REQ-ingest-market-data-scope, which the author has not reviewed) and raised this as question #7 on docs/implementation-questions.md. On 2026-10-31 the author chose a manifest declaration approved with the grant, as D-112 does for cross-check prices (REQ-ingest-cross-check-prices-declared-only), and recorded it as D-123; REQ-plugins-install-shows-market-data-list records the reason, that the install shows the owner the plugin will learn what they hold.

## Source

> Only a connector whose manifest declares it a market-data connector, and whose grant approves that, receives D-93's list of held instruments and model targets, so the install shows the owner that the plugin will learn what they hold.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-123#1
