---
title: The xtb plugin must serve ImportFile through the SDK with the XTB mapper, deliver nothing on Sync, and ask in its manifest for statement files only, with no egress and no Secret.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The xtb plugin must serve ImportFile through the SDK with the XTB mapper, deliver nothing on Sync, and ask in its manifest for statement files only, with no egress and no Secret.
logic_claims:
  - CLAIM-5357CCCD71EB167B
semantic_clauses:
  - The xtb plugin must serve ImportFile through the SDK with the XTB mapper, deliver nothing on Sync, and ask in its manifest for statement files only, with no egress and no Secret.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7011a57265b5272f753ec9b954de12d89eda44ff630076c0f81e84a99f3ed398
semantic_inventory:
  - claim_key: CLAIM-5357CCCD71EB167B
    claim_text: The xtb plugin must serve ImportFile through the SDK with the XTB mapper, deliver nothing on Sync, and ask in its manifest for statement files only, with no egress and no Secret
    role: normative
    status: modeled
    span:
      start: 0
      end: 177
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:13.405Z'
id: REQ-plugins-xtb-connector
type: req
---
The xtb plugin must serve ImportFile through the SDK with the XTB mapper, deliver nothing on Sync, and ask in its manifest for statement files only, with no egress and no Secret.

## Context

D-9 brings XTB in through uploaded files only, so the plugin needs no network or credential; D-109 shows the upload form for connectors that declare statement files. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> M0 data paths: XTB through uploaded xStation export files

docs/DESIGN.md#D-9
