---
title: The Polish pack's totals must offset the unrealized gains of every taxable account's listed holdings with the tax year's realized gains and losses, round the PIT-38 base and tax half up to whole złoty, keep bonds' own tax, tax an IKE's early return at 19% above what was paid in, tax IKZE on the whole account, and add OKI's rows as they are.
status: open
text_ref: docs/DESIGN.md#D-128
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-128
  - lane:strict
semantic_text: The Polish pack's totals must offset the unrealized gains of every taxable account's listed holdings with the tax year's realized gains and losses, round the PIT-38 base and tax half up to whole złoty, keep bonds' own tax, tax an IKE's early return at 19% above what was paid in, tax IKZE on the whole account, and add OKI's rows as they are.
logic_claims:
  - CLAIM-2D62F5F532F78FA4
semantic_clauses:
  - The Polish pack's totals must offset the unrealized gains of every taxable account's listed holdings with the tax year's realized gains and losses, round the PIT-38 base and tax half up to whole złoty, keep bonds' own tax, tax an IKE's early return at 19% above what was paid in, tax IKZE on the whole account, and add OKI's rows as they are.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 113ab3a68c536b5de95cf25aa042ddcd16cb2a32199cd3ac2ff41d4a13cfbbf7
semantic_inventory:
  - claim_key: CLAIM-2D62F5F532F78FA4
    claim_text: The Polish pack's totals must offset the unrealized gains of every taxable account's listed holdings with the tax year's realized gains and losses, round the PIT-38 base and tax half up to whole złoty, keep bonds' own tax, tax an IKE's early return at 19% above what was paid in, tax IKZE on the whole account, and add OKI's rows as they are
    role: normative
    status: modeled
    span:
      start: 0
      end: 342
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:01.929Z'
id: REQ-pl-totals-per-wrapper
type: req
---
The Polish pack's totals must offset the unrealized gains of every taxable account's listed holdings with the tax year's realized gains and losses, round the PIT-38 base and tax half up to whole złoty, keep bonds' own tax, tax an IKE's early return at 19% above what was paid in, tax IKZE on the whole account, and add OKI's rows as they are.

## Context

D-128 asks for PIT-38 netting on totals and IKE's own rule; whole-złoty rounding half up is art. 63 § 1 of the tax ordinance. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-128]** A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

docs/DESIGN.md#D-128
