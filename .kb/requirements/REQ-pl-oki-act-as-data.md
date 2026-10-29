---
title: The OKI act's in-force date, limit and rate must be data the pack holds, not in force while unknown, and OKI's yearly tax must be the rate of the part of the average daily balance above the limit, averaged over the year's days on or after the in-force date and rounded half up to grosze.
status: open
text_ref: docs/DESIGN.md#D-76
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-76
  - lane:strict
semantic_text: The OKI act's in-force date, limit and rate must be data the pack holds, not in force while unknown, and OKI's yearly tax must be the rate of the part of the average daily balance above the limit, averaged over the year's days on or after the in-force date and rounded half up to grosze.
logic_claims:
  - CLAIM-EF6B31632E5DF159
semantic_clauses:
  - The OKI act's in-force date, limit and rate must be data the pack holds, not in force while unknown, and OKI's yearly tax must be the rate of the part of the average daily balance above the limit, averaged over the year's days on or after the in-force date and rounded half up to grosze.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 34b957ca18e26c7a6fd940c40e7d5acbea1f891f55fa7edec8fb2ed85820f30e
semantic_inventory:
  - claim_key: CLAIM-EF6B31632E5DF159
    claim_text: The OKI act's in-force date, limit and rate must be data the pack holds, not in force while unknown, and OKI's yearly tax must be the rate of the part of the average daily balance above the limit, averaged over the year's days on or after the in-force date and rounded half up to grosze
    role: normative
    status: modeled
    span:
      start: 0
      end: 286
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:29.663Z'
id: REQ-pl-oki-act-as-data
type: req
---
The OKI act's in-force date, limit and rate must be data the pack holds, not in force while unknown, and OKI's yearly tax must be the rate of the part of the average daily balance above the limit, averaged over the year's days on or after the in-force date and rounded half up to grosze.

## Context

D-76 computes the yearly tax only once the act is in force, with its limit and rate. The act's numbers aren't known to the code, so the in-force date is unset and nothing is computed (docs/implementation-questions.md); averaging over the days given is this package's narrow reading. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-76]** Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

docs/DESIGN.md#D-76
