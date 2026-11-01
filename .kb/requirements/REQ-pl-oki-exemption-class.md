---
title: The Polish pack must classify a treasury savings bond as a savings asset, PLN cash as a savings asset on an account that held such bonds in the year and an investment asset otherwise, foreign cash as other, and report that it can't classify a holding priced at market, which counts as other.
status: open
text_ref: docs/DESIGN.md#D-132
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-132
  - lane:strict
semantic_text: The Polish pack must classify a treasury savings bond as a savings asset, PLN cash as a savings asset on an account that held such bonds in the year and an investment asset otherwise, foreign cash as other, and report that it can't classify a holding priced at market, which counts as other.
logic_claims:
  - CLAIM-26916DC17607483B
semantic_clauses:
  - The Polish pack must classify a treasury savings bond as a savings asset, PLN cash as a savings asset on an account that held such bonds in the year and an investment asset otherwise, foreign cash as other, and report that it can't classify a holding priced at market, which counts as other.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fba858674af4a7b23029b5493825c9ab6afa6e71c5b9b27c733d77188ccbaba2
semantic_inventory:
  - claim_key: CLAIM-26916DC17607483B
    claim_text: The Polish pack must classify a treasury savings bond as a savings asset, PLN cash as a savings asset on an account that held such bonds in the year and an investment asset otherwise, foreign cash as other, and report that it can't classify a holding priced at market, which counts as other
    role: normative
    status: modeled
    span:
      start: 0
      end: 290
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:54.519Z'
id: REQ-pl-oki-exemption-class
type: req
---
The Polish pack must classify a treasury savings bond as a savings asset, PLN cash as a savings asset on an account that held such bonds in the year and an investment asset otherwise, foreign cash as other, and report that it can't classify a holding priced at market, which counts as other.

## Context

D-127 names the classes and D-132 says what an instrument the pack can't classify counts as. The pack holds no data on listed instruments, so it can classify none of them. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-132]** When the pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, such as a fund whose policy it can't read, the instrument counts as neither, so the estimate errs high, and an attention item (D-130) asks the owner to classify it once; the owner's choice stays with the instrument.

docs/DESIGN.md#D-132
