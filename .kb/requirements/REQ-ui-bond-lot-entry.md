---
title: The built-in UI must offer an owner-only Add bond lot form on the page of an account at the treasury bond registry, hand each posted lot to the core's BondLotService with the plugin's own token, and show the core's reason with the lot as posted when it is refused.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The built-in UI must offer an owner-only Add bond lot form on the page of an account at the treasury bond registry, hand each posted lot to the core's BondLotService with the plugin's own token, and show the core's reason with the lot as posted when it is refused.
logic_claims:
  - CLAIM-F57FD99693648932
semantic_clauses:
  - The built-in UI must offer an owner-only Add bond lot form on the page of an account at the treasury bond registry, hand each posted lot to the core's BondLotService with the plugin's own token, and show the core's reason with the lot as posted when it is refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9552e7f23d33a6529e405a6c75efe31afa3c3d1fe71ef2b1e64c9856bd8c6099
semantic_inventory:
  - claim_key: CLAIM-F57FD99693648932
    claim_text: The built-in UI must offer an owner-only Add bond lot form on the page of an account at the treasury bond registry, hand each posted lot to the core's BondLotService with the plugin's own token, and show the core's reason with the lot as posted when it is refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 263
origin:
  kind: agent
  recorded_at: '2026-10-29T02:14:36.298Z'
id: REQ-ui-bond-lot-entry
type: req
---
The built-in UI must offer an owner-only Add bond lot form on the page of an account at the treasury bond registry, hand each posted lot to the core's BondLotService with the plugin's own token, and show the core's reason with the lot as posted when it is refused.

## Context

D-49: manual bond-lot entries in the built-in UI go to the core. Showing the form only for the institution treasury-bonds, on the account's page, is this task's narrow reading. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
