---
title: A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-223#1
tags:
  - views
  - period
  - performance
  - d-223
  - d-61
  - d-138
origin:
  kind: import
  ref: design-decisions:D-223#1
  recorded_at: '2026-11-15T18:11:37.441Z'
semantic_text: A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9aeca525b2983dd487e46eb404f5e49cc90dfc967c0828b51fa3ae98932148d6
logic_claims:
  - CLAIM-F8BA46C58CEF26A3
semantic_inventory:
  - claim_key: CLAIM-F8BA46C58CEF26A3
    claim_text: A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 135
    subjects:
      - views.performance
id: REQ-views-period-return-from-opening-close
type: req
pair_reviews:
  - with: REQ-portfolio-twr-cash-flows-end-of-day
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The time-weighted return must count cash flows at the end of their day; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-portfolio-twr-confirmed-transfers-only
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: Only confirmed transfers must be left out of the time-weighted return''s cash flows; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-portfolio-twr-opening-lots-inflows
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The time-weighted return must count opening lots as inflows at their market value on their date; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-portfolio-twr-rounding
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The time-weighted return must round each daily factor and the running product half-even to 12 decimal places; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-paid-out-payout-counts-as-return
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-performance-confirmed-transfers
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between th; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-performance-gaps
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The performance series must stop before the first day that is after an account''s valued-through date, on which a holding lacks one complete valuation of its current quant; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
  - with: REQ-views-performance-twr
    verdict: no_conflict
    separated_by: 'which close a period''s return starts from (this requirement) against: The read API''s performance series must chain each day''s return of the portfolio''s value with the day''s net external cash flow taken at the end of the day, so that deposit; this requirement changes none of the other''s rules and the other changes none of this one''s'
    claims:
      - claim_key: CLAIM-F8BA46C58CEF26A3
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
---
A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full.

## Context

The author decided it as D-223 on 2026-11-15, after the Ask box's period cards (REQ-ui-ask-every-period-one-card) put every period on one convention: 2025's return runs from the close of 31 December 2024. D-61 chains the time-weighted return across cash flows and D-138 puts the performance index at 100 on a chart's start, which had made a chart opened on a period's first day leave that day out. The core's GetPeriodChange computes it, from the opening day it names. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> **[D-223]** A period's return runs from the close of the day before its first day to the close of its last day, so its first day counts in full: 2025's runs from the close of 31 December 2024.

docs/DESIGN.md#D-223, D-61, D-138