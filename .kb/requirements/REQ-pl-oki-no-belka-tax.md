---
title: No Belka tax must be due inside an OKI account.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-76-3
  - lane:strict
text_ref: design-decisions:D-76#3
origin:
  kind: import
  ref: design-decisions:D-76#3
  recorded_at: '2026-10-29T07:55:22.542Z'
semantic_text: No Belka tax must be due inside an OKI account.
logic_claims:
  - CLAIM-CEF482CD21BD9F7C
semantic_clauses:
  - No Belka tax must be due inside an OKI account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c42f14631188e5e69b9a3d75db6a654b5513151da76b73d563e22c29ab389905
semantic_inventory:
  - claim_key: CLAIM-CEF482CD21BD9F7C
    claim_text: No Belka tax must be due inside an OKI account
    role: normative
    status: modeled
    span:
      start: 0
      end: 46
id: REQ-pl-oki-no-belka-tax
type: req
---
No Belka tax must be due inside an OKI account.

## Context

OKI (osobiste konto inwestycyjne) is one of the Polish pack's tax wrappers under D-76 (REQ-pl-tax-wrappers). R-18 records its draft rules: no Belka tax inside, and instead a yearly tax on the part of the average value above a limit, which REQ-pl-oki-yearly-tax records; sources disagree on whether the act is in force. The implementation already deducts no Belka tax inside IKE or IKZE (REQ-pl-tax-wrapper-exempt). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-76#3
