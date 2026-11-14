---
title: The owner can filter the audit log by outcome and by any of a set of actions as well as by actor, action and time, and an entry of a viewer's session names the person it is of by ID.
status: open
origin:
  kind: agent
  ref: 'S7 task brief: the Access page''s history and Settings'' audit log view (D-210, D-211), 2026-11-13'
  recorded_at: '2026-11-14T00:30:00.000Z'
firmness: default
semantic_text: The owner can filter the audit log by outcome and by any of a set of actions as well as by actor, action and time, and an entry of a viewer's session names the person it is of by ID.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 52d4ab4451a549e4532f9799c658718f8ea6a2ebeb1f030a9e5315221324708c
semantic_inventory:
  - claim_key: CLAIM-4C1FB5D39A083A82
    claim_text: The owner can filter the audit log by outcome and by any of a set of actions as well as by actor, action and time, and an entry of a viewer's session names the person it is of by ID
    role: descriptive
    status: missing
    span:
      start: 0
      end: 181
logic_claims:
  - CLAIM-4C1FB5D39A083A82
id: REQ-identity-audit-filter-outcome-actions
type: req
---
The owner can filter the audit log by outcome and by any of a set of actions as well as by actor, action and time, and an entry of a viewer's session names the person it is of by ID.

## Context

D-211 says Settings shows the log to the owner, filterable, and D-210 puts the history of every grant and revocation on the Access page. The pages need the outcome and a set of actions to filter by, and the viewer's person to name them: the log names a viewer's session, not their ID, so the core adds the person's ID, never a name.

## Source

S7 task brief: Settings shows it to the owner, filterable (D-211); the history of every grant and revocation (D-210).