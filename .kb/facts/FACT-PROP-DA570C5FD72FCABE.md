---
title: Delivers Due Coupons As Interest = each coupon due since its last sync, computed with the Polish pack's bond rules from the sync request's holdings and open lots, with no bond logic in the core
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-117-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-117#2
fact_kind: property_value
subject_key: plugins.treasury_bonds
property_key: delivers_due_coupons_as_interest
operator: eq
value_type: string
value_string: each coupon due since its last sync, computed with the Polish pack's bond rules from the sync request's holdings and open lots, with no bond logic in the core
canonical_key: kb-requirements-req-treasurybonds-delivers-coupons-md:plugins.treasury_bonds:delivers_due_coupons_as_interest:eq:each_coupon_due_since_its_last_sync_computed_with_the_polish_pack_s_bond_rules_from_the_sync_request_s_holdings_and_open_lots_with_no_bond_logic_in_the_core
claim_key: CLAIM-CDEAF7E0F8304C2C
claim_text: The treasury-bonds connector must deliver each coupon that fell due since its last sync as interest, computed with the Polish pack's bond rules from the holdings and open lots its sync request carries, so that the core holds no bond logic.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:12:11.670Z'
id: FACT-PROP-DA570C5FD72FCABE
type: fact
---
