---
title: The built-in UI must show the installed plugins with their capabilities, trust level and phase to owner and read sessions without any control that changes a plugin or grant, and must mark transactions from a removed instance in an account's transaction list.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: The built-in UI must show the installed plugins with their capabilities, trust level and phase to owner and read sessions without any control that changes a plugin or grant, and must mark transactions from a removed instance in an account's transaction list.
logic_claims:
  - CLAIM-B0C57926B6D73A67
semantic_clauses:
  - The built-in UI must show the installed plugins with their capabilities, trust level and phase to owner and read sessions without any control that changes a plugin or grant, and must mark transactions from a removed instance in an account's transaction list.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 22e98202a92e1f1ecf0cf22338360eeafc4c39deba4070230a9c8c43235d1997
semantic_inventory:
  - claim_key: CLAIM-B0C57926B6D73A67
    claim_text: The built-in UI must show the installed plugins with their capabilities, trust level and phase to owner and read sessions without any control that changes a plugin or grant, and must mark transactions from a removed instance in an account's transaction list
    role: normative
    status: modeled
    span:
      start: 0
      end: 257
origin:
  kind: agent
  recorded_at: '2026-10-29T04:02:41.506Z'
id: REQ-ui-installed-plugins
type: req
---
The built-in UI must show the installed plugins with their capabilities, trust level and phase to owner and read sessions without any control that changes a plugin or grant, and must mark transactions from a removed instance in an account's transaction list.

## Context

D-46: in M0 the UI only shows installed plugins; D-56: the UI shows the signature status; D-58: removed instances' data is marked. Read sessions see the list on /plugins because no decision restricts it; the owner also sees it on the settings page. Modeled by Claude Code for the plugin status task; the author has not reviewed it yet.

## Source

> **[D-46]** Plugins are installed with the `openfief` CLI. [...] The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md#D-46
