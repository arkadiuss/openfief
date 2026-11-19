---
title: 'The built-in UI''s pages report a fixed set of client events to the UI: Turnstile''s states and error codes with their timings, uncaught script errors as script, line and a short message from its own scripts only, failed requests by status, and the browser''s family and version; every field is validated against a fixed vocabulary and what does not fit is dropped, nothing typed, no amount, identifier, token or page content is ever reported, reports are size-capped and rate-limited per address, become log records without the address and OpenTelemetry metrics, it is on by default and can be turned off, and then the page sends nothing and the endpoint discards (D-227).'
status: open
firmness: firm
origin:
  kind: agent
  ref: design-decisions:D-227#1, the author's decision (docs/DESIGN.md)
  recorded_at: '2026-11-19T11:11:26.632Z'
id: REQ-ui-client-events
type: req
semantic_text: 'The built-in UI''s pages report a fixed set of client events to the UI: Turnstile''s states and error codes with their timings, uncaught script errors as script, line and a short message from its own scripts only, failed requests by status, and the browser''s family and version; every field is validated against a fixed vocabulary and what does not fit is dropped, nothing typed, no amount, identifier, token or page content is ever reported, reports are size-capped and rate-limited per address, become log records without the address and OpenTelemetry metrics, it is on by default and can be turned off, and then the page sends nothing and the endpoint discards (D-227).'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d740a60637eb7e854306e024a4139014d6322732a2c926b0c5a2a0164fb46f6d
semantic_inventory:
  - claim_key: CLAIM-DDE087F0E34353DE
    claim_text: 'The built-in UI''s pages report a fixed set of client events to the UI: Turnstile''s states and error codes with their timings, uncaught script errors as script, line and a short message from its own scripts only, failed requests by status, and the browser''s family and version'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 275
  - claim_key: CLAIM-97451A713DF3BAA7
    claim_text: every field is validated against a fixed vocabulary and what does not fit is dropped, nothing typed, no amount, identifier, token or page content is ever reported, reports are size-capped and rate-limited per address, become log records without the address and OpenTelemetry metrics, it is on by default and can be turned off, and then the page sends nothing and the endpoint discards (D-227)
    role: descriptive
    status: missing
    span:
      start: 277
      end: 669
logic_claims:
  - CLAIM-DDE087F0E34353DE
  - CLAIM-97451A713DF3BAA7
text_ref: design-decisions:D-227#1
required_evidence: asserted
---
The built-in UI's pages report a fixed set of client events to the UI: Turnstile's states and error codes with their timings, uncaught script errors as script, line and a short message from its own scripts only, failed requests by status, and the browser's family and version; every field is validated against a fixed vocabulary and what does not fit is dropped, nothing typed, no amount, identifier, token or page content is ever reported, reports are size-capped and rate-limited per address, become log records without the address and OpenTelemetry metrics, it is on by default and can be turned off, and then the page sends nothing and the endpoint discards (D-227).

## Context

Turnstile never completed on Brave for Android and the cause could not be seen from the server: the check runs in the browser. D-227 lets the UI's pages tell the UI what happened, as a fixed vocabulary, so that such problems are debugged from the event stream rather than by guessing, without ever reporting what a person typed or the record's data (GR-12). The owner's switch in Settings needs a core-stored setting; until it is built, the UI plugin's `clientEvents` config is the switch.

## Source

> D-227 (the author's decision, docs/DESIGN.md at commit 0668002): "The built-in UI reports a fixed set of client events to its own core: Turnstile's states and error codes, uncaught script errors as their message, script and line, failed requests by status, and the browser's family and version. It never reports anything typed, any amount, identifier or page content (GR-12). Reports are rate-limited per address and size-capped, become the core's log records and metrics, and leave the install only where OpenTelemetry export is set (D-172). It is on by default, and the owner can turn it off in Settings."