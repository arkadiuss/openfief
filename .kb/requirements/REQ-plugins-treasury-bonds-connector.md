---
title: The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach no external host, and answer a sync with an empty delivery.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach no external host, and answer a sync with an empty delivery.
logic_claims:
  - CLAIM-61EFE71E7D0FC4AF
semantic_clauses:
  - The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach no external host, and answer a sync with an empty delivery.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e06f3c906a632c52a62cbf4f3b4ca4bae7f7b2acc4b55d529dc76d1381f5dee4
semantic_inventory:
  - claim_key: CLAIM-61EFE71E7D0FC4AF
    claim_text: The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach no external host, and answer a sync with an empty delivery
    role: normative
    status: modeled
    span:
      start: 0
      end: 193
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:49.088Z'
id: REQ-plugins-treasury-bonds-connector
type: req
---
The treasury-bonds connector plugin must map bond lots the core hands it into the account its config names, serve only the core, reach no external host, and answer a sync with an empty delivery.

## Context

D-9 and GR-3 call for a treasury bonds connector; D-49 says the core hands manual entries to it; D-48 gives each instance its own config, so one instance serves one account. Neither the registry nor the bank offers an API, so it has no egress. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
