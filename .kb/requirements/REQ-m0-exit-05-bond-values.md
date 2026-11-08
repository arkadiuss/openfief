---
title: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; the liquidation value must match a hand calculation for at least one series of each type held; and totals must carry all three values.
status: open
text_ref: design-decisions:D-200#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-5
  - lane:strict
semantic_text: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; the liquidation value must match a hand calculation for at least one series of each type held; and totals must carry all three values.
logic_claims:
  - CLAIM-B5FBEA947A426D12
semantic_clauses:
  - For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; the liquidation value must match a hand calculation for at least one series of each type held; and totals must carry all three values.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e9db1068dc59cfa95980e8c2df1485f95c535a8ed05576ba0b96d2dc9eda3ef8
semantic_inventory:
  - claim_key: CLAIM-B5FBEA947A426D12
    claim_text: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; the liquidation value must match a hand calculation for at least one series of each type held; and totals must carry all three values
    role: normative
    status: modeled
    span:
      start: 0
      end: 444
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#5
  recorded_at: '2026-11-08T01:32:01.661Z'
id: REQ-m0-exit-05-bond-values
type: req
---
For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; the liquidation value must match a hand calculation for at least one series of each type held; and totals must carry all three values.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-5.

## Source

> **Bonds.** For every held series, the gross accrued value matches an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and per held series, with its formulas recorded in research, and the liquidation value matches a hand calculation for at least one series of each type held. Totals carry all three values (D-77).

design-decisions:D-200#5 (docs/DESIGN.md, D-200, criterion 5)

## Why this changed

REQ-m0-exit-05-bonds was imported from criterion 5's first wording, which named "net-if-redeemed" and had totals "use the configured basis, net by default". D-77 has no basis setting: every holding, account and total carries its gross, net and liquidation values side by side, tables show all three and the charts' switches start on gross, and the glossary retires net-if-redeemed for the liquidation value. The author reworded the criterion to match (R-29 records the hand calculation and the pinned reference valuer), so the requirement is imported again from the new text.
