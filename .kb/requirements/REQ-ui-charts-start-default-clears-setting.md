---
title: Going back to the default charts' start must clear the stored setting.
status: open
text_ref: design-decisions:D-167#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-167-2
  - lane:strict
semantic_text: Going back to the default charts' start must clear the stored setting.
logic_claims:
  - CLAIM-7F1AF744039FEC55
semantic_clauses:
  - Going back to the default charts' start must clear the stored setting.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a87b0d0f4434603f5000d428f6f655b4a7cb841695e99412306a7fad45d711d3
semantic_inventory:
  - claim_key: CLAIM-7F1AF744039FEC55
    claim_text: Going back to the default charts' start must clear the stored setting
    role: normative
    status: modeled
    span:
      start: 0
      end: 69
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-167#2
  recorded_at: '2026-11-02T00:56:03.667Z'
id: REQ-ui-charts-start-default-clears-setting
type: req
---
Going back to the default charts' start must clear the stored setting.

## Context

The author recorded it on 2026-11-01. Reason not stated. The defaults are D-138's: the portfolio's latest seeding from a starting balance, an account's own first day. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The charts' start the owner chooses (D-138) is a setting, stored for the portfolio and for each account, like the number format (D-113); going back to the default clears it.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-167#2
