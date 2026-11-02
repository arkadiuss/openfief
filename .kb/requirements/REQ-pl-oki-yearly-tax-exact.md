---
title: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, kept exact, exempting savings assets first then investment assets up to the year's amounts, rounding the rest and its tax at the year's rate to whole złoty, under 50 groszy down, and marking a year counted through a day before its last as an estimate, with no rate or tax while the minister hasn't announced the year's rate or exempt amounts.
status: open
text_ref: docs/DESIGN.md#D-144
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-144
  - lane:strict
semantic_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, kept exact, exempting savings assets first then investment assets up to the year's amounts, rounding the rest and its tax at the year's rate to whole złoty, under 50 groszy down, and marking a year counted through a day before its last as an estimate, with no rate or tax while the minister hasn't announced the year's rate or exempt amounts.
logic_claims:
  - CLAIM-917B5329BF20EBE4
semantic_clauses:
  - The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, kept exact, exempting savings assets first then investment assets up to the year's amounts, rounding the rest and its tax at the year's rate to whole złoty, under 50 groszy down, and marking a year counted through a day before its last as an estimate, with no rate or tax while the minister hasn't announced the year's rate or exempt amounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff760d9580ba383b3aecd6797ace6f55e804876ea3a0cf3dbc600a40996f54ec
semantic_inventory:
  - claim_key: CLAIM-917B5329BF20EBE4
    claim_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, kept exact, exempting savings assets first then investment assets up to the year's amounts, rounding the rest and its tax at the year's rate to whole złoty, under 50 groszy down, and marking a year counted through a day before its last as an estimate, with no rate or tax while the minister hasn't announced the year's rate or exempt amounts
    role: exception
    status: modeled
    span:
      start: 0
      end: 552
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:49.608Z'
id: REQ-pl-oki-yearly-tax-exact
type: req
---
The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, kept exact, exempting savings assets first then investment assets up to the year's amounts, rounding the rest and its tax at the year's rate to whole złoty, under 50 groszy down, and marking a year counted through a day before its last as an estimate, with no rate or tax while the minister hasn't announced the year's rate or exempt amounts.

## Context

D-144 keeps each account's average exact and rounds the base and the tax to whole złoty; D-145 takes later rates and exempt amounts only from the minister's announcements; D-146 makes the current year an estimate over its days so far. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Why this changed

It supersedes REQ-pl-oki-tax-computation: D-144 keeps each account's average exact and rounds the base and the tax to whole złoty instead of grosze, D-145 takes later rates and exempt amounts from the minister's announcements instead of computing them, and D-146 marks the current year an estimate. The superseded requirement is closed.

## Source

> **[D-144]** OKI's tax base and tax are rounded to whole złoty, as the tax ordinance rounds bases and taxes; each account's average value is kept exact.

docs/DESIGN.md#D-144
