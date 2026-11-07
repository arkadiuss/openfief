---
title: The Add bond lot form's help must say that the series' ISIN comes from the ministry's data with its rates, that a series not listed yet waits and shows in the Inbox after a week, and that a fee past the reviewed table is assumed for the owner to confirm.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: The Add bond lot form's help must say that the series' ISIN comes from the ministry's data with its rates, that a series not listed yet waits and shows in the Inbox after a week, and that a fee past the reviewed table is assumed for the owner to confirm.
logic_claims:
  - CLAIM-1243CF158AE690DA
semantic_clauses:
  - The Add bond lot form's help must say that the series' ISIN comes from the ministry's data with its rates, that a series not listed yet waits and shows in the Inbox after a week, and that a fee past the reviewed table is assumed for the owner to confirm.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f577f71eed00990b94a0cf09c98c3002485ab8619e4b27e1b0238b1531ea0d8b
semantic_inventory:
  - claim_key: CLAIM-1243CF158AE690DA
    claim_text: The Add bond lot form's help must say that the series' ISIN comes from the ministry's data with its rates, that a series not listed yet waits and shows in the Inbox after a week, and that a fee past the reviewed table is assumed for the owner to confirm
    role: normative
    status: modeled
    span:
      start: 0
      end: 253
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:40.769Z'
id: REQ-ui-bond-lot-help-isin
type: req
---
The Add bond lot form's help must say that the series' ISIN comes from the ministry's data with its rates, that a series not listed yet waits and shows in the Inbox after a week, and that a fee past the reviewed table is assumed for the owner to confirm.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. The design session asked that the add-lot help mention the assumed fee (D-97) with the next bonds change. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

docs/DESIGN.md#D-214
