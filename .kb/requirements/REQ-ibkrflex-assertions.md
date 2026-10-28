---
title: Each IBKR Flex statement must yield a holding balance assertion per stock or ETF from its open positions, using the SUMMARY row or else the sum of the LOT rows, and a cash balance assertion per currency from its cash report, all dated on the statement's last day.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Each IBKR Flex statement must yield a holding balance assertion per stock or ETF from its open positions, using the SUMMARY row or else the sum of the LOT rows, and a cash balance assertion per currency from its cash report, all dated on the statement's last day.
logic_claims:
  - CLAIM-4F29CB6236713CAB
semantic_clauses:
  - Each IBKR Flex statement must yield a holding balance assertion per stock or ETF from its open positions, using the SUMMARY row or else the sum of the LOT rows, and a cash balance assertion per currency from its cash report, all dated on the statement's last day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff32d3fe74484f288d42978b41dff8d40354f78be5652430d284f773aaeadcbc
semantic_inventory:
  - claim_key: CLAIM-4F29CB6236713CAB
    claim_text: Each IBKR Flex statement must yield a holding balance assertion per stock or ETF from its open positions, using the SUMMARY row or else the sum of the LOT rows, and a cash balance assertion per currency from its cash report, all dated on the statement's last day
    role: normative
    status: modeled
    span:
      start: 0
      end: 262
origin:
  kind: agent
  recorded_at: '2026-10-28T06:24:02.494Z'
id: REQ-ibkrflex-assertions
type: req
---
Each IBKR Flex statement must yield a holding balance assertion per stock or ETF from its open positions, using the SUMMARY row or else the sum of the LOT rows, and a cash balance assertion per currency from its cash report, all dated on the statement's last day.

## Context

D-18 stores broker position snapshots as balance assertions and flags a mismatch as an error. The Flex open positions and cash report are those snapshots; the base-currency summary row is a total, not cash in any currency, so it is not asserted. Refines REQ-AUTO-D9FBE33256704D04. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The Flex schema is IBKR's public Flex Query reference, not yet verified against a real account (OQ-22).

## Source

> Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
