---
title: Every change of the agents-and-models settings must be written to the audit log, naming the owner's session and the setting changed and no model list, inside the transaction that stores it, so that a change the log can't record is not made.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the MS1 brief, from D-211, 2026-11-13
  recorded_at: '2026-11-14T00:24:16.631Z'
semantic_text: Every change of the agents-and-models settings must be written to the audit log, naming the owner's session and the setting changed and no model list, inside the transaction that stores it, so that a change the log can't record is not made.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4ad9995d9fb6db8efe97217fbd06c3f8c00fd05309bf2826b33e2644b04079e9
semantic_inventory:
  - claim_key: CLAIM-7318C53CB4C7EC17
    claim_text: Every change of the agents-and-models settings must be written to the audit log, naming the owner's session and the setting changed and no model list, inside the transaction that stores it, so that a change the log can't record is not made
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 239
logic_claims:
  - CLAIM-7318C53CB4C7EC17
semantic_clauses:
  - Every change of the agents-and-models settings must be written to the audit log, naming the owner's session and the setting changed and no model list, inside the transaction that stores it, so that a change the log can't record is not made
id: REQ-models-settings-changes-audited
type: req
---
Every change of the agents-and-models settings must be written to the audit log, naming the owner's session and the setting changed and no model list, inside the transaction that stores it, so that a change the log can't record is not made.

## Context

M1 audited model calls and the maintainer's decisions but not the settings that allow them. A change of the switch, a toggle, a route or a tier's models is now an entry model_settings_changed, like the log's other entries failing closed.

## Source

> Settings: change each audited - check M1's audit coverage for settings changes, add if missing.

Source: the MS1 brief, from D-211, 2026-11-13
