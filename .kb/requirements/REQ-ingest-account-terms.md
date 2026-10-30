---
title: Creating an account must record its tax wrapper, the pack's default when none is given, and its fee rule, the zero rule when none is given, and changing an account must record a new wrapper or fee rule when one is given, each checked before anything is stored and answered with the account.
status: open
text_ref: docs/DESIGN.md#D-76
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-76
  - lane:strict
semantic_text: Creating an account must record its tax wrapper, the pack's default when none is given, and its fee rule, the zero rule when none is given, and changing an account must record a new wrapper or fee rule when one is given, each checked before anything is stored and answered with the account.
logic_claims:
  - CLAIM-92D04307D0C52597
semantic_clauses:
  - Creating an account must record its tax wrapper, the pack's default when none is given, and its fee rule, the zero rule when none is given, and changing an account must record a new wrapper or fee rule when one is given, each checked before anything is stored and answered with the account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 98bdbad96603d1f07ff7a71a8e2d75547744a36222654c2dad81ced045534767
semantic_inventory:
  - claim_key: CLAIM-92D04307D0C52597
    claim_text: Creating an account must record its tax wrapper, the pack's default when none is given, and its fee rule, the zero rule when none is given, and changing an account must record a new wrapper or fee rule when one is given, each checked before anything is stored and answered with the account
    role: normative
    status: modeled
    span:
      start: 0
      end: 289
origin:
  kind: agent
  recorded_at: '2026-10-29T22:59:09.979Z'
id: REQ-ingest-account-terms
type: req
---
Creating an account must record its tax wrapper, the pack's default when none is given, and its fee rule, the zero rule when none is given, and changing an account must record a new wrapper or fee rule when one is given, each checked before anything is stored and answered with the account.

## Context

D-76 has the owner set the wrapper when creating the account (D-72); D-78 makes the fee rule per account and zero until set. Letting the owner change both later, which revalues the account, is this package's reading. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-76]** Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

docs/DESIGN.md#D-76
