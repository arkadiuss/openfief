---
title: An attention item's outbox event must carry its subject as read access sees it (attention.ReadSubject), so no outbox row holds a broker account number; the item keeps its full subject for the owner.
status: open
text_ref: docs/DESIGN.md#D-121
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-121
  - lane:strict
semantic_text: An attention item's outbox event must carry its subject as read access sees it (attention.ReadSubject), so no outbox row holds a broker account number; the item keeps its full subject for the owner.
logic_claims:
  - CLAIM-C5256BA3B0044FAA
semantic_clauses:
  - An attention item's outbox event must carry its subject as read access sees it (attention.ReadSubject), so no outbox row holds a broker account number; the item keeps its full subject for the owner.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8e4cc165eebc45631177b632483c58bb5aee510e94219fe1ee78616f707fb392
semantic_inventory:
  - claim_key: CLAIM-C5256BA3B0044FAA
    claim_text: An attention item's outbox event must carry its subject as read access sees it (attention.ReadSubject), so no outbox row holds a broker account number; the item keeps its full subject for the owner
    role: normative
    status: modeled
    span:
      start: 0
      end: 197
origin:
  kind: agent
  recorded_at: '2026-11-03T15:54:57.511Z'
id: REQ-attention-outbox-read-subject
type: req
---
An attention item's outbox event must carry its subject as read access sees it (attention.ReadSubject), so no outbox row holds a broker account number; the item keeps its full subject for the owner.

## Context

D-121 keeps broker account numbers to the owner's account service, and D-165 has an item tied to no account name only its connector instance on the read API. The outbox's events are published beyond the owner's view, so the review before the alpha tag (P1f) found that the outbox must hold the same read form. The owner's view of the item is unchanged.

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

docs/DESIGN.md#D-121
