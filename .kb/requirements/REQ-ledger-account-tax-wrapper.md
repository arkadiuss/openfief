---
title: Every account must have a tax wrapper from the jurisdiction pack, which the owner sets when creating the account.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-76-1
  - lane:strict
text_ref: design-decisions:D-76#1
origin:
  kind: import
  ref: design-decisions:D-76#1
  recorded_at: '2026-10-29T07:53:58.382Z'
semantic_text: Every account must have a tax wrapper from the jurisdiction pack, which the owner sets when creating the account.
logic_claims:
  - CLAIM-55F8D50F8F5B5AF6
semantic_clauses:
  - Every account must have a tax wrapper from the jurisdiction pack, which the owner sets when creating the account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 37c2d37f15f2305e8c93c36981f263c9dc0659f48d7efa46cc76ee219935ae10
semantic_inventory:
  - claim_key: CLAIM-55F8D50F8F5B5AF6
    claim_text: Every account must have a tax wrapper from the jurisdiction pack, which the owner sets when creating the account
    role: normative
    status: modeled
    span:
      start: 0
      end: 112
id: REQ-ledger-account-tax-wrapper
type: req
---
Every account must have a tax wrapper from the jurisdiction pack, which the owner sets when creating the account.

## Context

D-72 has the owner create accounts in the built-in UI, in an owner session (REQ-ingest-owner-creates-accounts), and D-70 makes a jurisdiction pack a set of pure functions inside the core. The implementation already values bond lots in the account's tax wrapper, taxable unless recorded otherwise (REQ-valuation-bond-holding); D-76 makes the wrapper a property of every account that the owner sets. The author recorded D-76 to D-78 after the design session that answered the implementation's questions; D-77's three values are computed per tax wrapper.

## Source

> Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-76#1
