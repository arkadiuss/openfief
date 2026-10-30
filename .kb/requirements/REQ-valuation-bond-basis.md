---
title: A bond lot must count in totals at the configured valuation basis, gross_accrued or net_if_redeemed, with net_if_redeemed when nothing is configured, and the market basis must be rejected for retail treasury bonds.
status: closed
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: A bond lot must count in totals at the configured valuation basis, gross_accrued or net_if_redeemed, with net_if_redeemed when nothing is configured, and the market basis must be rejected for retail treasury bonds.
logic_claims:
  - CLAIM-10BC7640F4F926E3
semantic_clauses:
  - A bond lot must count in totals at the configured valuation basis, gross_accrued or net_if_redeemed, with net_if_redeemed when nothing is configured, and the market basis must be rejected for retail treasury bonds.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff810ada5090246d14f0db13bc20ce2de638d66b968f27bdf1577c7a905211f6
semantic_inventory:
  - claim_key: CLAIM-10BC7640F4F926E3
    claim_text: A bond lot must count in totals at the configured valuation basis, gross_accrued or net_if_redeemed, with net_if_redeemed when nothing is configured, and the market basis must be rejected for retail treasury bonds
    role: normative
    status: modeled
    span:
      start: 0
      end: 213
origin:
  kind: agent
  recorded_at: '2026-10-28T05:30:18.678Z'
id: REQ-valuation-bond-basis
type: req
---
A bond lot must count in totals at the configured valuation basis, gross_accrued or net_if_redeemed, with net_if_redeemed when nothing is configured, and the market basis must be rejected for retail treasury bonds.

## Context

Closed: superseded by REQ-valuation-bond-holding-values. D-77 removed the configured bond basis; gross, net and liquidation replace it. Modeled by Claude Code for the values work package; the author has not reviewed it yet.
