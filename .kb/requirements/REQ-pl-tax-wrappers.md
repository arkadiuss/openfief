---
title: The Polish jurisdiction pack's tax wrappers must be a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne).
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-76-2
  - lane:strict
text_ref: design-decisions:D-76#2
origin:
  kind: import
  ref: design-decisions:D-76#2
  recorded_at: '2026-10-29T07:54:40.404Z'
semantic_text: The Polish jurisdiction pack's tax wrappers must be a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne).
logic_claims:
  - CLAIM-F747945D0FE8EE6B
semantic_clauses:
  - The Polish jurisdiction pack's tax wrappers must be a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne).
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 505b2216d5788af3ee8771bc14e27ce2e22d6a1f9eb9b898834866fec441fe45
semantic_inventory:
  - claim_key: CLAIM-F747945D0FE8EE6B
    claim_text: The Polish jurisdiction pack's tax wrappers must be a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne)
    role: normative
    status: modeled
    span:
      start: 0
      end: 127
id: REQ-pl-tax-wrappers
type: req
---
The Polish jurisdiction pack's tax wrappers must be a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne).

## Context

D-76 gives every account a tax wrapper from the jurisdiction pack (REQ-ledger-account-tax-wrapper). The implementation already records IKE and IKZE wrappers for bond valuation (REQ-pl-tax-wrapper-exempt); D-76 names the standard taxable account and adds OKI, the planned Polish personal investment account whose draft rules and uncertain status R-18 records. The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-76#2
