---
title: The built-in UI must keep the charts' start the owner chose in the page's URL, as a view parameter shared by the page's time charts, with a link back to the default, and must not store it as a setting.
status: closed
text_ref: docs/DESIGN.md#D-138
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-138
  - lane:strict
semantic_text: The built-in UI must keep the charts' start the owner chose in the page's URL, as a view parameter shared by the page's time charts, with a link back to the default, and must not store it as a setting.
logic_claims:
  - CLAIM-12495E0FCF793B91
semantic_clauses:
  - The built-in UI must keep the charts' start the owner chose in the page's URL, as a view parameter shared by the page's time charts, with a link back to the default, and must not store it as a setting.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c26edfe365229332650f6d0d6be49496a63f723544696c1e8087649bf7102a2
semantic_inventory:
  - claim_key: CLAIM-12495E0FCF793B91
    claim_text: The built-in UI must keep the charts' start the owner chose in the page's URL, as a view parameter shared by the page's time charts, with a link back to the default, and must not store it as a setting
    role: normative
    status: modeled
    span:
      start: 0
      end: 200
origin:
  kind: agent
  recorded_at: '2026-11-01T13:02:05.536Z'
id: REQ-ui-chart-start-view-parameter
type: req
---
The built-in UI must keep the charts' start the owner chose in the page's URL, as a view parameter shared by the page's time charts, with a link back to the default, and must not store it as a setting.

## Context

D-138 says the owner can choose the date the charts start on but, unlike D-113's number format, doesn't call it a setting, so keeping it in the URL is this implementation's reading (asked as implementation question 56). Recorded by the implementation agent of impl/chart-start for D-138; default firmness, not reviewed by the author yet.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

docs/DESIGN.md#D-138
