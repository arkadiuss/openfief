---
title: The check of incomplete bond series must raise one item of kind bond-series-incomplete per ISIN a series is held under, among the decisions, for a series held seven days or more since any account first held it under its provisional ISIN or without terms, saying what is missing, and stay silent before.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: The check of incomplete bond series must raise one item of kind bond-series-incomplete per ISIN a series is held under, among the decisions, for a series held seven days or more since any account first held it under its provisional ISIN or without terms, saying what is missing, and stay silent before.
logic_claims:
  - CLAIM-68CD200DCDCF37B5
semantic_clauses:
  - The check of incomplete bond series must raise one item of kind bond-series-incomplete per ISIN a series is held under, among the decisions, for a series held seven days or more since any account first held it under its provisional ISIN or without terms, saying what is missing, and stay silent before.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5f32a0599327af35459a7361c42a9c61850676e2fafd7de430fc2402e299debc
semantic_inventory:
  - claim_key: CLAIM-68CD200DCDCF37B5
    claim_text: The check of incomplete bond series must raise one item of kind bond-series-incomplete per ISIN a series is held under, among the decisions, for a series held seven days or more since any account first held it under its provisional ISIN or without terms, saying what is missing, and stay silent before
    role: normative
    status: modeled
    span:
      start: 0
      end: 301
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:36.968Z'
id: REQ-attention-bond-series-incomplete
type: req
---
The check of incomplete bond series must raise one item of kind bond-series-incomplete per ISIN a series is held under, among the decisions, for a series held seven days or more since any account first held it under its provisional ISIN or without terms, saying what is missing, and stay silent before.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. The design session read the seven days as silence while the ministry's data is unavailable, and the item as a decision tied to the instrument (D-130, D-196). Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

docs/DESIGN.md#D-214
