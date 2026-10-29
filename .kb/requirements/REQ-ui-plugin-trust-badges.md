---
title: The built-in UI must group each installed plugin's signature trust into one of four badges, Verified, Yours, Unreviewed or Blocked, and show the full signature trust and the exact signer with it.
status: open
text_ref: docs/DESIGN.md#D-87
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-87
  - lane:strict
semantic_text: The built-in UI must group each installed plugin's signature trust into one of four badges, Verified, Yours, Unreviewed or Blocked, and show the full signature trust and the exact signer with it.
logic_claims:
  - CLAIM-5E320C6029DE2F9E
semantic_clauses:
  - The built-in UI must group each installed plugin's signature trust into one of four badges, Verified, Yours, Unreviewed or Blocked, and show the full signature trust and the exact signer with it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e039547113d1afa65a0f92b8c06c98745a0f1d3462dfb1e3c1ca37cb0c8ffa36
semantic_inventory:
  - claim_key: CLAIM-5E320C6029DE2F9E
    claim_text: The built-in UI must group each installed plugin's signature trust into one of four badges, Verified, Yours, Unreviewed or Blocked, and show the full signature trust and the exact signer with it
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-10-29T09:50:01.644Z'
id: REQ-ui-plugin-trust-badges
type: req
---
The built-in UI must group each installed plugin's signature trust into one of four badges, Verified, Yours, Unreviewed or Blocked, and show the full signature trust and the exact signer with it.

## Context

D-87 is the author's decision on how the UI shows plugin trust. Treating OpenFief's CI and catalog publishers outside the catalog as Unreviewed, and an unknown trust as Blocked, is the UI agent's literal reading.

## Source

> The UI groups them into four badges, with the exact signer in the details: Verified (in the catalog and reviewed, D-44), Yours (one of the owner's signers, D-45), Unreviewed (another trusted signer, not in the catalog) and Blocked (refused or not checked yet).

docs/DESIGN.md#D-87
