---
title: The core stores the owner's switch for the built-in UI's client events, on until the owner turns it off, lets only the owner change it, enters each change in the audit log as a settings change that names the setting and no value, and answers the view plugin's read of it without a session.
status: open
firmness: default
required_evidence: unverified
semantic_text: The core stores the owner's switch for the built-in UI's client events, on until the owner turns it off, lets only the owner change it, enters each change in the audit log as a settings change that names the setting and no value, and answers the view plugin's read of it without a session.
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
tags:
  - client-events
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: daaa189861fc985becc93ccdf6f32b0692060f4f6cace9ecf5481cddc0def0a4
logic_claims:
  - CLAIM-11EC50157C140375
semantic_inventory:
  - claim_key: CLAIM-11EC50157C140375
    claim_text: The core stores the owner's switch for the built-in UI's client events, on until the owner turns it off, lets only the owner change it, enters each change in the audit log as a settings change that names the setting and no value, and answers the view plugin's read of it without a session
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 288
    subjects:
      - settings.client_events_switch
    reason: It states how the built-in UI behaves; no predicate of the ontology states it yet.
rationale: D-227 promises that the owner can turn the client events off in Settings, and D-211 audits every settings change. The read needs no session because the login page reports too and the UI asks before anyone has signed in; it carries one boolean.
id: REQ-settings-client-events-switch
type: req
---
The core stores the owner's switch for the built-in UI's client events, on until the owner turns it off, lets only the owner change it, enters each change in the audit log as a settings change that names the setting and no value, and answers the view plugin's read of it without a session.

## Context

D-227 promises that the owner can turn the client events off in Settings, and D-211 audits every settings change. The read needs no session because the login page reports too and the UI asks before anyone has signed in; it carries one boolean.

## Source

> D-227 (docs/DESIGN.md): It is on by default, and the owner can turn it off in Settings. Reports [...] become the core's log records and metrics, and leave the install only where OpenTelemetry export is set (D-172).
