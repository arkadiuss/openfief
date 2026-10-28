---
title: The catalog and the owner's trusted-signer list must be parsed strictly, catalog publishers must have exact identities, and a document with any invalid entry or unknown field must be refused as a whole.
status: open
text_ref: docs/DESIGN.md#D-43
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-43
  - lane:strict
semantic_text: The catalog and the owner's trusted-signer list must be parsed strictly, catalog publishers must have exact identities, and a document with any invalid entry or unknown field must be refused as a whole.
logic_claims:
  - CLAIM-D6129F0F94907060
semantic_clauses:
  - The catalog and the owner's trusted-signer list must be parsed strictly, catalog publishers must have exact identities, and a document with any invalid entry or unknown field must be refused as a whole.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 45f5360967822f6965506e4103c0a787f7e8622f89e2e678c1f65a0e4f0cc89b
semantic_inventory:
  - claim_key: CLAIM-D6129F0F94907060
    claim_text: The catalog and the owner's trusted-signer list must be parsed strictly, catalog publishers must have exact identities, and a document with any invalid entry or unknown field must be refused as a whole
    role: normative
    status: modeled
    span:
      start: 0
      end: 201
origin:
  kind: agent
  recorded_at: '2026-10-28T08:52:23.159Z'
id: REQ-plugins-catalog-format
type: req
---
The catalog and the owner's trusted-signer list must be parsed strictly, catalog publishers must have exact identities, and a document with any invalid entry or unknown field must be refused as a whole.

## Context

D-43 makes the catalog a file in this repository; D-44 has each entry pin a digest and the publisher's signer identity, which this reads as an exact subject or key, never a pattern; D-45 lets the owner add signers. A trust document that is half-read could silently widen trust, so any problem refuses all of it. The format holds only what the trust check needs; the CLI's install flow (D-46) may extend it. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> One catalog, a file in this repository released with the product, lists the first-party and verified plugins.

docs/DESIGN.md D-43

> its catalog entry pins the publisher's signed image digest and the publisher's signer identity.

docs/DESIGN.md D-44
