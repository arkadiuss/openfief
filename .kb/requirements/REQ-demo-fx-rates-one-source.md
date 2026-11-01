---
title: The demo loader must record its FX rates as from the demo's NBP connector instance, so each currency's rates have one source that the instance's syncs continue.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The demo loader must record its FX rates as from the demo's NBP connector instance, so each currency's rates have one source that the instance's syncs continue.
logic_claims:
  - CLAIM-97ABEA86711B1E48
semantic_clauses:
  - The demo loader must record its FX rates as from the demo's NBP connector instance, so each currency's rates have one source that the instance's syncs continue.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7b74ccb80d150d61298937a817b4491ef918ac76aaf88674d5d9ffbbf2dd42c1
semantic_inventory:
  - claim_key: CLAIM-97ABEA86711B1E48
    claim_text: The demo loader must record its FX rates as from the demo's NBP connector instance, so each currency's rates have one source that the instance's syncs continue
    role: normative
    status: modeled
    span:
      start: 0
      end: 159
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:35.823Z'
id: REQ-demo-fx-rates-one-source
type: req
---
The demo loader must record its FX rates as from the demo's NBP connector instance, so each currency's rates have one source that the instance's syncs continue.

## Context

Before D-94 was rewritten, the loader's rates (source nbp-a) and the live instance's filled each other's days. With one source per series that per-date fallback is gone, so the loader records its rates under the instance's name. Claude Code's choice for the demo; not reviewed by the author.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
