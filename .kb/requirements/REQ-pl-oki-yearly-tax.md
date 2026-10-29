---
title: OKI's yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, must be computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-76-4
  - lane:strict
text_ref: design-decisions:D-76#4
origin:
  kind: import
  ref: design-decisions:D-76#4
  recorded_at: '2026-10-29T07:56:13.505Z'
semantic_text: OKI's yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, must be computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.
logic_claims:
  - CLAIM-0DBEEEAE958D7393
semantic_clauses:
  - OKI's yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, must be computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c5b2dba9c408b4f03357fa0b5358e72ab37b851d09e83b54392ee9d016dcca1a
semantic_inventory:
  - claim_key: CLAIM-0DBEEEAE958D7393
    claim_text: OKI's yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, must be computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act
    role: normative
    status: modeled
    span:
      start: 0
      end: 245
id: REQ-pl-oki-yearly-tax
type: req
---
OKI's yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, must be computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

## Context

OKI (osobiste konto inwestycyjne) is one of the Polish pack's tax wrappers under D-76 (REQ-pl-tax-wrappers), and no Belka tax is due inside it (REQ-pl-oki-no-belka-tax). R-18 records the draft rules: a yearly tax on the part of the average value of qualified assets above 100,000 PLN, summed across all of a taxpayer's OKI accounts. Sources disagree on that tax's rate and on whether the act is in force; D-76 takes the limit and rate from the act and computes the tax only once it is in force. Stored daily valuations are D-36's (REQ-decision-d36-1-stored-valuations). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-76#4
