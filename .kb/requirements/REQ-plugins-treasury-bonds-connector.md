---
title: The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach only GUS's public API and NBP's published archive, and answer a sync with the index readings after its watermark.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach only GUS's public API and NBP's published archive, and answer a sync with the index readings after its watermark.
logic_claims:
  - CLAIM-53F6EBD54B1F3EC7
semantic_clauses:
  - The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach only GUS's public API and NBP's published archive, and answer a sync with the index readings after its watermark.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ba4d8b65f228da82bcc24a62f0a77a7f36c4a5630fbe812e6ba5a3292ca36e8c
semantic_inventory:
  - claim_key: CLAIM-53F6EBD54B1F3EC7
    claim_text: The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach only GUS's public API and NBP's published archive, and answer a sync with the index readings after its watermark
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:49.088Z'
id: REQ-plugins-treasury-bonds-connector
type: req
---
The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach only GUS's public API and NBP's published archive, and answer a sync with the index readings after its watermark.

## Context

D-9 and GR-3 call for a treasury bonds connector; D-49 says the core hands manual entries to it; D-48 gives each instance its own config, so one instance serves one account. Neither the registry nor the bank offers an API (R-5), so lots are entered by hand; CPI and the reference rate are public, so the plugin needs egress but no secret. Modeled by Claude Code for the bond indices task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
