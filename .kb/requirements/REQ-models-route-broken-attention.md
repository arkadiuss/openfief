---
title: A route entry naming a provider instance that no plugin has, or a plugin that is not a model provider, falls through to the route's next model and raises one attention item among the Inbox's decisions for that entry, naming the route and the instance, which clears when the owner fixes the route; a provider that is only slow or down falls through and raises none.
status: open
firmness: default
origin:
  kind: agent
  ref: design-session answers for the M1 merge and follow-ups, 2026-11-13 (D-196, D-205)
  recorded_at: '2026-11-13T14:45:45.612Z'
semantic_text: A route entry naming a provider instance that no plugin has, or a plugin that is not a model provider, falls through to the route's next model and raises one attention item among the Inbox's decisions for that entry, naming the route and the instance, which clears when the owner fixes the route; a provider that is only slow or down falls through and raises none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d25cb3bce3213fbeeb8eac6a4f90eb0df3ddc4fd7b6c38af63101b3ae30d9591
semantic_inventory:
  - claim_key: CLAIM-0FD5F511BC1DC313
    claim_text: A route entry naming a provider instance that no plugin has, or a plugin that is not a model provider, falls through to the route's next model and raises one attention item among the Inbox's decisions for that entry, naming the route and the instance, which clears when the owner fixes the route
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 295
  - claim_key: CLAIM-08FBF01217B56FD9
    claim_text: a provider that is only slow or down falls through and raises none
    role: descriptive
    status: missing
    span:
      start: 297
      end: 363
logic_claims:
  - CLAIM-0FD5F511BC1DC313
  - CLAIM-08FBF01217B56FD9
id: REQ-models-route-broken-attention
type: req
---
A route entry naming a provider instance that no plugin has, or a plugin that is not a model provider, falls through to the route's next model and raises one attention item among the Inbox's decisions for that entry, naming the route and the instance, which clears when the owner fixes the route; a provider that is only slow or down falls through and raises none.

## Context

D-196 groups the Inbox's decisions, among them missing settings. A route that names an instance the owner never installed is a setting to fix, unlike a timeout, which fixes itself. The design session that answered the M1 follow-ups asked for one item per broken entry and none for a transient failure, which D-208's guards will count later.

## Source

> Route naming a missing / non-model-provider instance: fall through AND raise an attention item in Inbox group decisions -> missing settings (D-196), one per broken route entry naming route + instance, cleared when fixed. Transient provider failure (timeout/5xx): fall through, no item (count for D-208 later).

Source: design-session answers for the M1 merge and follow-ups, 2026-11-13