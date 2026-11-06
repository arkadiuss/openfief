---
title: The attention store must run a check whose conditions are the instruments priced at market held in the tax year an account in a wrapper with a yearly tax is valued through, which the pack can't classify and the owner hasn't, one per ISIN, so the item resolves once the owner classifies the instrument.
status: open
text_ref: docs/DESIGN.md#D-148
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-148
  - lane:strict
semantic_text: The attention store must run a check whose conditions are the instruments priced at market held in the tax year an account in a wrapper with a yearly tax is valued through, which the pack can't classify and the owner hasn't, one per ISIN, so the item resolves once the owner classifies the instrument.
logic_claims:
  - CLAIM-B6F9E9F631DC9798
semantic_clauses:
  - The attention store must run a check whose conditions are the instruments priced at market held in the tax year an account in a wrapper with a yearly tax is valued through, which the pack can't classify and the owner hasn't, one per ISIN, so the item resolves once the owner classifies the instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c48f58879a7d9688a4e9cdde0cacb2162b34f515abe11f2d7d4101ff7988adee
semantic_inventory:
  - claim_key: CLAIM-B6F9E9F631DC9798
    claim_text: The attention store must run a check whose conditions are the instruments priced at market held in the tax year an account in a wrapper with a yearly tax is valued through, which the pack can't classify and the owner hasn't, one per ISIN, so the item resolves once the owner classifies the instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 300
origin:
  kind: agent
  recorded_at: '2026-11-01T20:12:16.036Z'
id: REQ-attention-unclassified-instrument-check
type: req
---
The attention store must run a check whose conditions are the instruments priced at market held in the tax year an account in a wrapper with a yearly tax is valued through, which the pack can't classify and the owner hasn't, one per ISIN, so the item resolves once the owner classifies the instrument.

## Context

D-148 makes the unclassified instrument's item of D-132 a check the attention store runs. Looking at the tax year the account is valued through is a narrow reading. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Why this changed

It supersedes REQ-valuation-unclassified-attention: D-148 makes the unclassified instrument's item a check the attention store runs, so it resolves itself, instead of an item a revaluation raised and cleared through the attention port. The superseded requirement is closed.

## Source

> **[D-148]** An unclassified instrument's attention item (D-132) is a check the attention store runs, so it resolves itself once the owner classifies the instrument.

docs/DESIGN.md#D-148
