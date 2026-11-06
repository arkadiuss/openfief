---
title: The charts' start the owner chooses must be a setting, stored for the portfolio and for each account.
status: open
text_ref: design-decisions:D-167#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-167-1
  - lane:strict
semantic_text: The charts' start the owner chooses must be a setting, stored for the portfolio and for each account.
logic_claims:
  - CLAIM-0D7991537A6C60D9
semantic_clauses:
  - The charts' start the owner chooses must be a setting, stored for the portfolio and for each account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3ad12dfabaa97f630c85b5ebbd5c41a6fbafd52d33b49269a13804b23f60a3df
semantic_inventory:
  - claim_key: CLAIM-0D7991537A6C60D9
    claim_text: The charts' start the owner chooses must be a setting, stored for the portfolio and for each account
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-167#1
  recorded_at: '2026-11-02T00:55:57.916Z'
id: REQ-ui-charts-start-stored-setting
type: req
---
The charts' start the owner chooses must be a setting, stored for the portfolio and for each account.

## Why this changed

On 2026-11-01 the author recorded D-167 (reason not stated). REQ-ui-chart-start-view-parameter, a default requirement the implementation agent recorded for D-138, kept the chosen start in the page's URL and stored nothing, which D-167 contradicts, so this requirement replaces it, and it is closed. Not built yet: the code that keeps the start in the URL stays linked to the closed requirement until the setting is built.

## Context

The author recorded it on 2026-11-01. Reason not stated. The owner chooses the start under D-138; the decision likens the setting to the number format (D-113). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The charts' start the owner chooses (D-138) is a setting, stored for the portfolio and for each account, like the number format (D-113); going back to the default clears it.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-167#1
