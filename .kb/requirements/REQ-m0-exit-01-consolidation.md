---
title: The author's real IBKR, XTB stock and ETF and treasury bond holdings must appear in one view, every broker-reported position and cash balance must be stored as a balance assertion that matches the derived value to 0.01 of the account currency, and open XTB CFDs must appear only as flagged, unsupported items.
status: open
text_ref: design-decisions:D-200#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-1
  - lane:strict
semantic_text: The author's real IBKR, XTB stock and ETF and treasury bond holdings must appear in one view, every broker-reported position and cash balance must be stored as a balance assertion that matches the derived value to 0.01 of the account currency, and open XTB CFDs must appear only as flagged, unsupported items.
logic_claims:
  - CLAIM-77B947DE5C854A0E
semantic_clauses:
  - The author's real IBKR, XTB stock and ETF and treasury bond holdings must appear in one view, every broker-reported position and cash balance must be stored as a balance assertion that matches the derived value to 0.01 of the account currency, and open XTB CFDs must appear only as flagged, unsupported items.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4fd3409f4e753ba19213dbf1e767d434f10a3961730259776199c6a2a924113d
semantic_inventory:
  - claim_key: CLAIM-77B947DE5C854A0E
    claim_text: The author's real IBKR, XTB stock and ETF and treasury bond holdings must appear in one view, every broker-reported position and cash balance must be stored as a balance assertion that matches the derived value to 0.01 of the account currency, and open XTB CFDs must appear only as flagged, unsupported items
    role: normative
    status: modeled
    span:
      start: 0
      end: 308
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#1
  recorded_at: '2026-11-07T03:31:26.791Z'
id: REQ-m0-exit-01-consolidation
type: req
---
The author's real IBKR, XTB stock and ETF and treasury bond holdings must appear in one view, every broker-reported position and cash balance must be stored as a balance assertion that matches the derived value to 0.01 of the account currency, and open XTB CFDs must appear only as flagged, unsupported items.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-1. This criterion is the exception D-200 names: the owner checks it on a private install with real accounts, since no real data enters tests or CI (GR-12). Its verification is manual, by the author, and it gets no CI scenario; it is to be linked to the `openfief check balances` work, which checks the stored balance assertions against the derived values, once that lands.

## Source

> **Consolidation.** The author's real IBKR, XTB (stocks and ETFs) and treasury bond holdings appear in one view. Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.

design-decisions:D-200#1 (docs/DESIGN.md, D-200, criterion 1)
