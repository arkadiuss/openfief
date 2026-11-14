---
title: The built-in UI must give the owner alone a Models page in Settings, where the one switch for agents and models and a toggle for each feature are set, with the MCP server's toggle marked as not built yet, and every other session must get 403 and no link to it.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the MS1 brief, from D-212, 2026-11-13
  recorded_at: '2026-11-14T00:24:09.122Z'
semantic_text: The built-in UI must give the owner alone a Models page in Settings, where the one switch for agents and models and a toggle for each feature are set, with the MCP server's toggle marked as not built yet, and every other session must get 403 and no link to it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9100e141759f9cf65a836115e2ce98a35901167776840e933a94fab7e41cee64
semantic_inventory:
  - claim_key: CLAIM-AE399C52F3F57F11
    claim_text: The built-in UI must give the owner alone a Models page in Settings, where the one switch for agents and models and a toggle for each feature are set, with the MCP server's toggle marked as not built yet, and every other session must get 403 and no link to it
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 259
logic_claims:
  - CLAIM-AE399C52F3F57F11
semantic_clauses:
  - The built-in UI must give the owner alone a Models page in Settings, where the one switch for agents and models and a toggle for each feature are set, with the MCP server's toggle marked as not built yet, and every other session must get 403 and no link to it
id: REQ-ui-models-page-owner-only
type: req
---
The built-in UI must give the owner alone a Models page in Settings, where the one switch for agents and models and a toggle for each feature are set, with the MCP server's toggle marked as not built yet, and every other session must get 403 and no link to it.

## Context

D-212 puts the switch and its toggles in Settings; the core's ModelSettingsService existed with no page. The owner alone changes them, so a viewer or the demo never sees the page or its link. The MCP server's toggle is stored only, and the page says so plainly.

## Source

> Agents and models are off by default, behind one switch in Settings, with a toggle under it for each feature: the Ask box's model (D-203), the maintainer (D-204) and the MCP server (D-213).

Source: the MS1 brief, from D-212, 2026-11-13
