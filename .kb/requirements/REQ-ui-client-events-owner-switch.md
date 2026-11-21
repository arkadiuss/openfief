---
title: 'The built-in UI follows the owner''s client events setting: with it off the pages load no reporting script and the endpoint answers 204 and keeps nothing, with it on both work, the last answer is kept while the core can''t be asked and off when there is none, and Settings shows the switch with one line that says a fixed list of events is collected, nothing typed and no amounts, and that it stays on the install unless OpenTelemetry export is set.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The built-in UI follows the owner''s client events setting: with it off the pages load no reporting script and the endpoint answers 204 and keeps nothing, with it on both work, the last answer is kept while the core can''t be asked and off when there is none, and Settings shows the switch with one line that says a fixed list of events is collected, nothing typed and no amounts, and that it stays on the install unless OpenTelemetry export is set.'
origin:
  kind: agent
  ref: 'D-227 (docs/DESIGN.md): the owner can turn client events off in Settings; the author approved the deferral of the end-to-end test of /client-events (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
tags:
  - client-events
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c4215c74bccb83b4d1881ba9080bfe13c52ea0b43ac34588680bffd8fa15cd0d
logic_claims:
  - CLAIM-D2D2DA46B43E39C7
semantic_inventory:
  - claim_key: CLAIM-D2D2DA46B43E39C7
    claim_text: 'The built-in UI follows the owner''s client events setting: with it off the pages load no reporting script and the endpoint answers 204 and keeps nothing, with it on both work, the last answer is kept while the core can''t be asked and off when there is none, and Settings shows the switch with one line that says a fixed list of events is collected, nothing typed and no amounts, and that it stays on the install unless OpenTelemetry export is set'
    role: exception
    status: ontology_gap
    span:
      start: 0
      end: 446
    subjects:
      - ui.client_events_switch
    reason: It states how the built-in UI behaves; no predicate of the ontology states it yet.
rationale: D-227 says the owner can turn the client events off in Settings and that reports stay on the install unless OpenTelemetry export is set (D-172). The endpoint answers 204 and drops, as it did with the plugin's config, because pages opened before the change go on posting and a 404 would fill their consoles with errors.
id: REQ-ui-client-events-owner-switch
type: req
---
The built-in UI follows the owner's client events setting: with it off the pages load no reporting script and the endpoint answers 204 and keeps nothing, with it on both work, the last answer is kept while the core can't be asked and off when there is none, and Settings shows the switch with one line that says a fixed list of events is collected, nothing typed and no amounts, and that it stays on the install unless OpenTelemetry export is set.

## Context

D-227 says the owner can turn the client events off in Settings and that reports stay on the install unless OpenTelemetry export is set (D-172). The endpoint answers 204 and drops, as it did with the plugin's config, because pages opened before the change go on posting and a 404 would fill their consoles with errors.

## Source

> D-227 (docs/DESIGN.md): It is on by default, and the owner can turn it off in Settings. Reports [...] become the core's log records and metrics, and leave the install only where OpenTelemetry export is set (D-172).
