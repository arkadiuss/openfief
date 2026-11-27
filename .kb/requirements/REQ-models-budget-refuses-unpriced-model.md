---
title: With a daily budget set, the model router must refuse a call before sending it when it can't price the call's model, because the catalog has no entry for the model or its entry has no price although a third party runs it, with an error the Ask box shows as a failed precondition, recorded as an audit entry of class budget, and must never count such a call as free; a model on the owner's hardware with no price stays free.
status: open
firmness: default
semantic_text: With a daily budget set, the model router must refuse a call before sending it when it can't price the call's model, because the catalog has no entry for the model or its entry has no price although a third party runs it, with an error the Ask box shows as a failed precondition, recorded as an audit entry of class budget, and must never count such a call as free; a model on the owner's hardware with no price stays free.
origin:
  kind: agent
  ref: With --model-daily-budget-usd set, refuse a candidate without a catalog entry (and an entry with no price for a third_party_cloud model), with a typed error the UI can show. Test it with an empty catalog. (the owner's items for alpha.2, from the reviews of main 0bfeb62)
  recorded_at: '2026-11-27T08:04:40.685Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4bf26a6187b9b5e2b12734b31164552915aa4862f81ca656daffd5628e5d7699
semantic_inventory:
  - claim_key: CLAIM-A6DBE556D9421151
    claim_text: With a daily budget set, the model router must refuse a call before sending it when it can't price the call's model, because the catalog has no entry for the model or its entry has no price although a third party runs it, with an error the Ask box shows as a failed precondition, recorded as an audit entry of class budget, and must never count such a call as free
    role: normative
    status: missing
    span:
      start: 0
      end: 364
    subjects:
      - models.router_budget_pricing
  - claim_key: CLAIM-637C2C1D20450F22
    claim_text: a model on the owner's hardware with no price stays free
    role: descriptive
    status: missing
    span:
      start: 366
      end: 422
logic_claims:
  - CLAIM-A6DBE556D9421151
  - CLAIM-637C2C1D20450F22
id: REQ-models-budget-refuses-unpriced-model
type: req
---
With a daily budget set, the model router must refuse a call before sending it when it can't price the call's model, because the catalog has no entry for the model or its entry has no price although a third party runs it, with an error the Ask box shows as a failed precondition, recorded as an audit entry of class budget, and must never count such a call as free; a model on the owner's hardware with no price stays free.

## Context

The catalog is empty after a restart until ListModels succeeds, and the router skipped a candidate with no entry and booked its call at USD 0, so the day's first calls spent without being counted against the cap (finding S3 of the reviews of main 0bfeb62). The cap fails closed, as the rest of the budget does (REQ-models-daily-budget).

## Source

> With --model-daily-budget-usd set, refuse a candidate without a catalog entry (and an entry with no price for a third_party_cloud model), with a typed error the UI can show. Test it with an empty catalog. (the owner's items for alpha.2, from the reviews of main 0bfeb62)