---
title: The read API's installed plugins and the built-in UI's plugins page must show the institutions each connector's manifest declares it books for.
status: open
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
semantic_text: The read API's installed plugins and the built-in UI's plugins page must show the institutions each connector's manifest declares it books for.
logic_claims:
  - CLAIM-0BC4F37995FD374A
semantic_clauses:
  - The read API's installed plugins and the built-in UI's plugins page must show the institutions each connector's manifest declares it books for.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5a7525a9e1da095f6e0523eb46019b55672e2417bb31fdf486cd349f259d12ef
semantic_inventory:
  - claim_key: CLAIM-0BC4F37995FD374A
    claim_text: The read API's installed plugins and the built-in UI's plugins page must show the institutions each connector's manifest declares it books for
    role: normative
    status: modeled
    span:
      start: 0
      end: 142
origin:
  kind: agent
  recorded_at: '2026-11-02T22:20:11.330Z'
id: REQ-views-plugin-institutions
type: req
---
The read API's installed plugins and the built-in UI's plugins page must show the institutions each connector's manifest declares it books for.

## Context

So the owner, and anyone reading what runs (D-86), sees which institution's data each connector may deliver, alongside its signature trust. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-182 existed in kibi on this branch when it was built.

## Source

> A plugin's Secrets are named so they can never collide with the platform's, and a connector books only for the institutions its signed manifest declares.

docs/DESIGN.md#D-182
