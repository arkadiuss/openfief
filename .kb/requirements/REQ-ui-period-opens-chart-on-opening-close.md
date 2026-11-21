---
title: Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-223#2
tags:
  - ui
  - performance
  - period
  - d-223
  - d-138
  - d-61
origin:
  kind: import
  ref: design-decisions:D-223#2
  recorded_at: '2026-11-15T18:11:38.733Z'
semantic_text: Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dc89ce03b35413b70d739d4560d76c0c453af4d55d8e3b4d1a87ed9e468ad7b7
logic_claims:
  - CLAIM-F5005365120F0539
semantic_inventory:
  - claim_key: CLAIM-F5005365120F0539
    claim_text: Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 242
    subjects:
      - views.performance
      - views.time_charts
id: REQ-ui-period-opens-chart-on-opening-close
type: req
pair_reviews:
  - with: REQ-portfolio-twr-cash-flows-end-of-day
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The time-weighted return must count cash flows at the end of their day; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-portfolio-twr-confirmed-transfers-only
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: Only confirmed transfers must be left out of the time-weighted return''s cash flows; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-portfolio-twr-opening-lots-inflows
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The time-weighted return must count opening lots as inflows at their market value on their date; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-portfolio-twr-rounding
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The time-weighted return must round each daily factor and the running product half-even to 12 decimal places; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-account-chart-default-start
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: An account''s time charts must by default start on the account''s own first day; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-charts-start-default-clears-setting
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: Going back to the default charts'' start must clear the stored setting; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-charts-start-stored-setting
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The charts'' start the owner chooses must be a setting, stored for the portfolio and for each account; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-opened-account-keeps-default-start
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: An account marked as opened on its seeding day must not move the portfolio''s default charts'' start; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-performance-index-100-on-start
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The performance index must be 100 on a time chart''s start date; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-portfolio-chart-default-start
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The portfolio''s charts must by default start on the latest day before which some account''s history is unknown because it was seeded from a starting balance; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-portfolio-chart-start-ignores-full-history
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: Accounts with their whole history, or opened later, must not move the portfolio charts'' default start; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-time-chart-marks-tracking-start
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: A time chart must mark where tracking began; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-time-chart-start-owner-chosen
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: Every time chart, the portfolio''s and each account''s, must start on a date the owner can choose; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-ui-unmarked-seeded-account-moves-start
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: An account seeded from a starting balance and not marked as opened on its seeding day must move the portfolio''s default charts'' start; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-paid-out-payout-counts-as-return
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-performance-confirmed-transfers
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between th; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-performance-gaps
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The performance series must stop before the first day that is after an account''s valued-through date, on which a holding lacks one complete valuation of its current quant; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-performance-twr
    verdict: no_conflict
    separated_by: 'where a chart''s start is set when a period opens on it (this requirement) against: The read API''s performance series must chain each day''s return of the portfolio''s value with the day''s net external cash flow taken at the end of the day, so that deposit; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F5005365120F0539
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-period-return-from-opening-close
    verdict: no_conflict
    separated_by: where a chart's start is set when a period opens on it (the chart requirement) against which close a period's return starts from (the return requirement); both name the period's opening close and agree on it, and neither changes the other's rule
    claims:
      - claim_key: CLAIM-F5005365120F0539
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
---
Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return.

## Context

The author decided it as D-223 on 2026-11-15. D-138 and REQ-ui-performance-index-100-on-start keep the performance index at 100 on the charts' start, so a chart that starts on a period's first day shows a return that leaves that day's move out, and differs from the period card's, which runs from the close of the day before. Setting the start to the opening close, rather than changing the chart's convention, makes the two agree. A start the owner types by hand stays as typed. The overview has no preset row: the one link that opens a period on the chart besides the Ask box's card is the value card's time-weighted return of this year, which opens at the year's opening close. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> **[D-223]** Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start (D-138) is set to the period's opening close, so the chart, which keeps 100 on its start, and the card show the same return.

docs/DESIGN.md#D-223, D-138, D-61